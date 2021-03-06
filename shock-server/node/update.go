package node

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"os"
	"strconv"
	"strings"
	"time"
)

//Modification functions
func (node *Node) Update(params map[string]string, files FormFiles) (err error) {
	// Exclusive conditions
	// 1. has files[upload] (regular upload)
	// 2. has params[parts] (partial upload support)
	// 3. has params[type] & params[source] (v_node)
	// 4. has params[path] (set from local path)
	//
	// All condition allow setting of attributes

	if _, uploadMisplaced := params["upload"]; uploadMisplaced {
		return errors.New("upload form field must be file encoded.")
	}

	_, isRegularUpload := files["upload"]
	_, isPartialUpload := params["parts"]

	isVirtualNode := false
	if t, hasType := params["type"]; hasType && t == "virtual" {
		isVirtualNode = true
	}
	_, isPathUpload := params["path"]

	// Check exclusive conditions
	if (isRegularUpload && isPartialUpload) || (isRegularUpload && isVirtualNode) || (isRegularUpload && isPathUpload) {
		return errors.New("upload parameter incompatible with parts, path and/or type parameter(s)")
	} else if (isPartialUpload && isVirtualNode) || (isPartialUpload && isPathUpload) {
		return errors.New("parts parameter incompatible with type and/or path parameter(s)")
	} else if isVirtualNode && isPathUpload {
		return errors.New("type parameter incompatible with path parameter")
	}

	// Check if immutable
	if (isRegularUpload || isPartialUpload || isVirtualNode || isPathUpload) && node.HasFile() {
		return errors.New(e.FileImut)
	}

	if isRegularUpload {
		if err = node.SetFile(files["upload"]); err != nil {
			return err
		}
		delete(files, "upload")
	} else if isPartialUpload {
		if node.isVarLen() || node.partsCount() > 0 {
			return errors.New("parts already set")
		}
		// Number of parts should be either a positive integer or string 'unknown'
		if params["parts"] == "unknown" {
			if err = node.initParts("unknown"); err != nil {
				return err
			}
		} else if params["parts"] == "close" {
			if err = node.closeVarLenPartial(); err != nil {
				return err
			}
		} else {
			n, err := strconv.Atoi(params["parts"])
			if err != nil {
				return errors.New("parts must be an integer or 'unknown'")
			}
			if n < 1 {
				return errors.New("parts cannot be less than 1")
			}
			if err = node.initParts(params["parts"]); err != nil {
				return err
			}
		}
	} else if isVirtualNode {
		if source, hasSource := params["source"]; hasSource {
			ids := strings.Split(source, ",")
			node.addVirtualParts(ids)
		} else {
			return errors.New("type virtual requires source parameter")
		}
	} else if isPathUpload {
		localpaths := strings.Split(conf.Conf["local-paths"], ",")
		if len(localpaths) > 0 {
			for _, p := range localpaths {
				if strings.HasPrefix(params["path"], p) {
					if err = node.SetFileFromPath(params["path"]); err != nil {
						return err
					} else {
						return nil
					}
				}
			}
			return errors.New("file not in local files path. Please contact your Shock administrator.")
		} else {
			return errors.New("local files path uploads must be configured. Please contact your Shock administrator.")
		}
	}

	// set attributes from file
	if _, hasAttr := files["attributes"]; hasAttr {
		if err = node.SetAttributes(files["attributes"]); err != nil {
			return err
		}
		os.Remove(files["attributes"].Path)
		delete(files, "attributes")
	}

	// handle part file
	LockMgr.LockPartOp()
	parts_count := node.partsCount()
	if parts_count > 0 || node.isVarLen() {
		for key, file := range files {
			if node.HasFile() {
				LockMgr.UnlockPartOp()
				return errors.New(e.FileImut)
			}
			keyn, errf := strconv.Atoi(key)
			if errf == nil && (keyn <= parts_count || node.isVarLen()) {
				err = node.addPart(keyn-1, &file)
				if err != nil {
					LockMgr.UnlockPartOp()
					return err
				}
			} else {
				LockMgr.UnlockPartOp()
				return errors.New("invalid file parameter")
			}
		}
	}
	LockMgr.UnlockPartOp()

	// update relatives
	if _, hasRelation := params["linkage"]; hasRelation {
		ltype := params["linkage"]

		if ltype == "parent" {
			if node.HasParent() {
				return errors.New(e.ProvenanceImut)
			}
		}
		var ids string
		if _, hasIds := params["ids"]; hasIds {
			ids = params["ids"]
		} else {
			return errors.New("missing ids for updating relatives")
		}
		var operation string
		if _, hasOp := params["operation"]; hasOp {
			operation = params["operation"]
		}
		if err = node.UpdateLinkages(ltype, ids, operation); err != nil {
			return err
		}
	}

	//update node tags
	if _, hasDataType := params["tags"]; hasDataType {
		if err = node.UpdateDataTags(params["tags"]); err != nil {
			return err
		}
	}

	//update file format
	if _, hasFormat := params["format"]; hasFormat {
		if node.File.Format != "" {
			return errors.New(fmt.Sprintf("file format already set:%s", node.File.Format))
		}
		if err = node.SetFileFormat(params["format"]); err != nil {
			return err
		}
	}
	return
}

func (node *Node) Save() (err error) {
	node.UpdateVersion()
	if len(node.Revisions) == 0 || node.Revisions[len(node.Revisions)-1].Version != node.Version {
		n := Node{node.Id, node.Version, node.File, node.Attributes, node.Public, node.Indexes, node.Acl, node.VersionParts, node.Tags, nil, node.Linkages, node.CreatedOn, node.LastModified}
		node.Revisions = append(node.Revisions, n)
	}
	if node.CreatedOn == "" {
		node.CreatedOn = time.Now().Format(time.UnixDate)
	} else {
		node.LastModified = time.Now().Format(time.UnixDate)
	}

	bsonPath := fmt.Sprintf("%s/%s.bson", node.Path(), node.Id)
	os.Remove(bsonPath)
	nbson, err := bson.Marshal(node)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(bsonPath, nbson, 0644)
	if err != nil {
		return
	}
	err = dbUpsert(node)
	if err != nil {
		return
	}
	return
}

func (node *Node) UpdateVersion() (err error) {
	parts := make(map[string]string)
	h := md5.New()
	version := node.Id
	for name, value := range map[string]interface{}{"file_ver": node.File, "attributes_ver": node.Attributes, "acl_ver": node.Acl} {
		m, er := json.Marshal(value)
		if er != nil {
			return
		}
		h.Write(m)
		sum := fmt.Sprintf("%x", h.Sum(nil))
		parts[name] = sum
		version = version + ":" + sum
		h.Reset()
	}
	h.Write([]byte(version))
	node.Version = fmt.Sprintf("%x", h.Sum(nil))
	node.VersionParts = parts
	return
}

func (node *Node) UpdateLinkages(ltype string, ids string, operation string) (err error) {
	var link linkage
	link.Type = ltype
	idList := strings.Split(ids, ",")
	for _, id := range idList {
		link.Ids = append(link.Ids, id)
	}
	link.Operation = operation
	node.Linkages = append(node.Linkages, link)
	err = node.Save()
	return
}

func (node *Node) UpdateDataTags(types string) (err error) {
	tagslist := strings.Split(types, ",")
	for _, newtag := range tagslist {
		if contains(node.Tags, newtag) {
			continue
		}
		node.Tags = append(node.Tags, newtag)
	}
	err = node.Save()
	return
}
