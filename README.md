![Shock](https://raw.github.com/jaredwilkening/Shock/master/shock-server/site/assets/img/shock_logo.png)
=====

About:
------

Shock is a platform to support computation, storage, and distribution. Designed from the ground up to be fast, scalable, fault tolerant, federated. 

Shock is RESTful. Accessible from desktops, HPC systems, exotic hardware, the cloud and your smartphone.

Shock is for scientific data. One of the challenges of large volume scientific data is that without often complex metadata it is of little to no value. Store and query(in development) complex metadata.   

Shock is an data management system. Annotate, anonymize, convert, filter, quality control, statically subsample at line speed bioinformatics sequence data. Extensible plug-in architecture(in development).

**Most importantly Shock is still very much in development. Be patient and contribute.**

Shock is actively being developed at [github.com/MG-RAST/Shock](http://github.com/MG-RAST/Shock).

<br>
Building:
---------
Shock (requires go=>1.1.0 [golang.org](http://golang.org/), git, mercurial, bazaar):

    go get github.com/MG-RAST/Shock/...

Built binary will be located in env configured $GOPATH or $GOROOT depending on Go configuration.

<br>
Configuration:
--------------
The Shock configuration file is in INI file format. There is a template of the config file located at the root level of the repository.

    [Admin]
    email=admin@host.com
    secretkey=supersecretkey

    [Anonymous]
    # Controls an anonymous user's ability to read/write
    # values: true/false
    read=true
    write=false
    create-user=false

    [Auth]
    # defaults to local user management with basis auth
    type=basic
    # comment line about and uncomment below to use Globus Online as auth provider
    #type=globus 
    #globus_token_url=https://nexus.api.globusonline.org/goauth/token?grant_type=client_credentials
    #globus_profile_url=https://nexus.api.globusonline.org/users

    [Directories]
    # See documentation for details of deploying Shock
    site=/usr/local/shock/site
    data=/usr/local/shock
    logs=/var/log/shock
    
    # Comma delimited search path available for remote path uploads. Only remote paths that prefix 
    # match one of the following will be allowed. Note: poor choices can result in security concerns.
    local_paths=N/A

    [External]
    site-url=http://localhost
    
    [SSL]
    enable=false
    #key=<path_to_key_file>
    #cert=<path_to_cert_file>

    [Mongodb]
    # Mongodb configuration:
    # Hostnames and ports hosts=host1[,host2:port,...,hostN]
    hosts=localhost
    database=ShockDB
    user=
    password=

    [Mongodb-Node-Indices]
    # See http://www.mongodb.org/display/DOCS/Indexes#Indexes-CreationOptions for more info on mongodb index options.
    # key=unique:true/false[,dropDups:true/false][,sparse:true/false]
    id=unique:true

    [Ports]
    # Ports for site/api
    # Note: use of port 80 may require root access
    site-port=7444
    api-port=7445


<a name="auth"/>
<br>

Running:
--------------
To run (additional requires mongodb=>2.0.3):
  
    shock-server -conf <path_to_config_file>

<br>

Routes Overview
---------------
    
### API Routes (default port 7445):

#####OPTIONS

- all options request respond with CORS headers and 200 OK

#####GET

- [/](#get_slash)  resource listing
- [/node](#get_nodes)  list nodes, query
- [/node/{id}](#get_node)  view node, download file (full or partial)
- [/node/{id}/acl]()  view node acls
- [/node/{id}/acl/{type}]()  view node acls of type {type}

#####PUT

- [/node/{id}](#put_node)  modify node, create index
- [/node/{id}/acl]()  modify node acls
- [/node/{id}/acl/{type}]()  modify node acls of type {type}

#####POST
 
- [/node](#post_node)  create node

#####DELETE

- [/node/{id}]()  delete node

### Site Routes (default port 7444):

#####GET

- [/]() this documentation and future site
- [/raw]()    listing of data dir
- [/assets]() js, img, css, README.md

<br>

Authentication:
---------------
Shock supports multiple forms of Authentication via plugin modules. Credentials are cached for 1 hour to speed up high transaction loads. Server restarts will clear the credential cache.

### Globus Online 
In this configuration Shock locally stores only uuids for users that it has already seen. The registration of new users is done exclusively with the external auth provider. The user api is disabled in this mode.

Examples:

    # globus online username & password
    curl --user username:password ...

    # globus online bearer token 
    curl -H "Authorization: OAuth $TOKEN" ...


<br>

Data Types
----------

### Node:

- id: unique identifier
- file: name, size, checksum(s).
- attributes: arbitrary json. Queriable.
- indexes: A set of indexes to use.
- version: a version stamp for this node.

##### node example:
    
    {
        "data": {
            "attributes": null, 
            "file": {
                "checksum": {}, 
                "format": "", 
                "name": "", 
                "size": 0, 
                "virtual": false, 
                "virtual_parts": []
            }, 
            "id": "130cadb5-9435-4bd9-be13-715ec40b2bb5", 
            "relatives": [], 
            "type": [], 
            "version": "4da883924aa8ae9eb95f6cd247f2f554"
        }, 
        "error": null, 
        "status": 200
    }

<br>
### Index:

Currently there is support for two types of indices: virtual (for both size and record) and bam. 

##### virtual index:

A virtual index is one that can be generated on the fly without support of precalculated data. The current working example of this 
is the size virtual index. Based on the file size and desired chunksize the partitions become individually addressable. 

##### bam index (bai):

To use the bam index feature, the <a href="http://samtools.sourceforge.net/">SAMtools</a> package must be installed on the machine that is running the Shock server with the samtools executable in the path of the user that is running the Shock server.

Also, in order to use this feature, you must sort your .bam file using the 'samtools sort' command before uploading the file into Shock.

<br><br>

API by example
-----------------
All examples use curl but can be easily modified for any http client / library. 
__Note__: Authentication is required for most of these commands
<br>
#### Node Creation ([details](#post_node)):

    # without file or attributes
    curl -X POST http://<host>[:<port>]/node

    # with attributes
    curl -X POST -F "attributes=@<path_to_json_file>" http://<host>[:<port>]/node
    
    # with file
    curl -X POST -F "upload=@<path_to_data_file>" http://<host>[:<port>]/node

    # with file local to the shock server
    curl -X POST -F "path=<path_to_data_file>" http://<host>[:<port>]/node
    
    # with file upload in N parts (part uploads may be done in parallel)
    curl -X POST -F "parts=N" http://<host>[:<port>]/node
    curl -X PUT -F "1=@<file_part_1>" http://<host>[:<port>]/node/<node_id>
    curl -X PUT -F "2=@<file_part_2>" http://<host>[:<port>]/node/<node_id>
    ...
    curl -X PUT -F "N=@<file_part_N>" http://<host>[:<port>]/node/<node_id>

<br>
#### Node retrieval ([details](#get_node)):

    # node information
    curl -X GET http://<host>[:<port>]/node/{id}

    # download file
    curl -X GET http://<host>[:<port>]/node/{id}/?download

    # download first 1mb of file
    curl -X GET http://<host>[:<port>]/node/{id}/?download&index=size&part=1
        
    # download first 10mb of file
    curl -X GET http://<host>[:<port>]/node/{id}/?download&index=size&chunk_size=10485760&part=1

    # download Nth 10mb of file
    curl -X GET http://<host>[:<port>]/node/{id}/?download&index=size&chunk_size=10485760&part=N

    # download entire bam file in human readable sam alignments
    curl -X GET http://<host>[:<port>]/node/{id}/?download&index=bai

    # download bam alignments overlapped with specified region (ref_id:start_pos-end_pos)
    curl -X GET http://<host>[:<port>]/node/{id}/?download&index=bai&region=chr1:1-20000

    # download bam alignments with selected arguments supported by "samtools view"
    curl -X GET http://<host>[:<port>]/node/{id}/?download/index=bai&head&headonly&count&flag=[INT]&lib=[STR]&mapq=[INT]&readgroup=[STR]
    (note: all the arguments are optional and can be used with or without the region, but the index=bai is required)
    
<br>
#### Node acls: 

    # view all acls
    curl -X GET http://<host>[:<port>]/node/{id}/acl/

    # view specific acls
    curl -X GET http://<host>[:<port>]/node/{id}/acl/[ read | write | delete | owner ]

    # changing owner (chown)
    curl -X PUT http://<host>[:<port>]/node/{id}/acl/owner?users=<email-address_or_uuid>

    # adding user to all acls (expect owner)
    curl -X PUT http://<host>[:<port>]/node/{id}/acl/?all=<list_of_email-addresses_or_uuids>

    # adding user to specific acls
    curl -X PUT http://<host>[:<port>]/node/{id}/acl/[ read | write | delete | owner ]?users=<list_of_email-addresses_or_uuids>
    or
    curl -X PUT http://<host>[:<port>]/node/{id}/acl/?[ read | write | delete ]=<list_of_email-addresses_or_uuids>
    
    # adding users to both read and write acls:
    curl -X PUT http://<host>[:<port>]/node/{id}/acl/?read=<list_of_email-addresses_or_uuids>&write=<list_of_email-addresses_or_uuids>
    
    # deleting user from all acls (expect owner)
    curl -X DELETE http://<host>[:<port>]/node/{id}/acl/?all=<list_of_email-addresses_or_uuids>    
    
    # deleting user to specific acls
    curl -X DELETE http://<host>[:<port>]/node/{id}/acl/[ read | write | delete ]?users=<list_of_email-addresses_or_uuids>
    or
    curl -X DELETE http://<host>[:<port>]/node/{id}/acl/?[ read | write | delete ]=<list_of_email-addresses_or_uuids>
    
    # deleting users to both read and write acls:
    curl -X DELETE http://<host>[:<port>]/node/{id}/acl/?read=<list_of_email-addresses_or_uuids>&write=<list_of_email-addresses_or_uuids>

<br>
#### Querying ([details](#get_nodes)):

    # by attribute key value
    curl -X GET http://<host>[:<port>]/node/?query&<key>=<value>

    # by attribute key value, limit 10
    curl -X GET http://<host>[:<port>]/node/?query&<key>=<value>&limit=10

    # by attribute key value, limit 10, offset 10
    curl -X GET http://<host>[:<port>]/node/?query&<key>=<value>&limit=10&offset=10

<br>

API
---

### Response wrapper:
All responses from Shock currently are in the following encoding. 

  {
    "data": <JSON or null>,
    "error": <string or null: error message>,
    "status": <int: http status code>,
    "limit": <int: paginated requests only>, 
    "offset": <int: paginated requests only>,
    "total_count": <int: paginated requests only>
  }

<a name="get_slash"/>
<br>
### GET /

Description of resources available through this api

##### example
	
    curl -X GET http://<host>[:<port>]/
	
##### returns

    {"resources":["node", "user"],"url":"http://localhost:8000/","documentation":"http://localhost/","contact":"admin@host.com","id":"Shock","type":"Shock"}

<a name="post_node"/>
<br>
### POST /node

Create node

 - optionally takes user/password via Basic Auth. If set only that user with have access to the node
 - accepts multipart/form-data encoded 
 - to set attributes include file field named "attributes" containing a json file of attributes
 - to set file include file field named "upload" containing any file **or** include field named "path" containing the file system path to the file accessible from the Shock server

##### example
	
	curl -X POST [ see Authentication ] [ -F "attributes=@<path_to_json>" ( -F "upload=@<path_to_data_file>" || -F "path=<path_to_file>") ] http://<host>[:<port>]/node
	
##### returns

    {
        "data": {<node>},
        "error": <error message or null>, 
        "status": <http status of response (also set in headers)>
    } 

<a name="get_nodes"/>
<br>
### GET /node

List nodes

 - optionally takes user/password via Basic Auth. Grants access to non-public data
 - by adding ?offset=N you get the nodes starting at N+1 
 - by adding ?limit=N you get a maximum of N nodes returned 

##### querying
All attributes are queriable. For example if a node has in it's attributes "about" : "metagenome" the url 

    /node/?query&about=metagenome
    
would return it and all other nodes with that attribute. Address of nested attributes like "metadata": { "env_biome": "ENVO:human-associated habitat", ... } is done via a dot notation 

    /node/?query&metadata.env_biome=ENVO:human-associated%20habitat

Multiple attributes can be selected in a single query and are treated as AND operations

    /node/?query&metadata.env_biome=ENVO:human-associated%20habitat&about=metagenome
    
**Note:** all special characters like a space must be url encoded.

##### example
	
	curl -X GET [ see Authentication ] http://<host>[:<port>]/node/[?offset=<offset>&limit=<count>][&query&<tag>=<value>]
		
##### returns

    {
      "data": {[<array of nodes>]},
      "error": <string or null: error message>,
      "status": <int: http status code>,
      "limit": <limit>, 
      "offset": <offset>,
      "total_count": <count>
    }

<a name="get_node"/>
<br>
### GET /node/{id}

View node, download file (full or partial)

 - optionally takes user/password via Basic Auth
 - ?download - complete file download
 - ?download&index=size&part=1\[&part=2...\]\[chunksize=inbytes\] - download portion of the file via the size virtual index. Chunksize defaults to 1MB (1048576 bytes).

##### example	

	curl -X GET [ see Authentication ] http://<host>[:<port>]/node/{id}

##### returns

    {
        "data": {<node>},
        "error": <error message or null>, 
        "status": <http status of request>
    }

<a name="put_node"/>
<br>
### PUT /node/{id}

Modify node, create index

 - optionally takes user/password via Basic Auth
 
**Modify:** 

 - **Once the file or attributes of a node are set they are immutiable.**
 - accepts multipart/form-data encoded 
 - to set attributes include file field named "attributes" containing a json file of attributes
 - to set file include file field named "upload" containing any file **or** include field named "path" containing the file system path to the file accessible from the Shock server
   
##### example	
  
	curl -X PUT [ see Authentication ] [ -F "attributes=@<path_to_json>" ( -F "upload=@<path_to_data_file>" || -F "path=<path_to_file>") ] http://<host>[:<port>]/node/{id}

  
##### returns

    {
        "data": {<node>},
        "error": <error message or null>, 
        "status": <http status of request>
    }

<br>
**Create index:**

 - currently available index types: size, record (for sequence file types), and bai (bam index)

##### example	

	curl -X PUT [ see Authentication ] http://<host>[:<port>]/node/{id}?index=<type>

##### returns

    {
        "data": null,
        "error": <error message or null>, 
        "status": <http status of request>
    }

##### bam index (bai) argument mapping from URL to samtools

<table border=1>
    <tr>
        <td><b>URL argument</b></td>
        <td><b>value type</b></td>
        <td><b>samtools argument</b></td>
        <td><b>operation</b></td>
    </tr>
    <tr>
        <td>head</td>
        <td>no value</td>
        <td>-h</td>
        <td>Include the header in the output</td>
    </tr>
    <tr>
        <td>headonly</td>
        <td>no value</td>
        <td>-H</td>
        <td>Output the header only.</td>
    </tr>
    <tr>
        <td>count</td>
        <td>no value</td>
        <td>-c</td>
        <td>Instead of printing the alignments, only count them and print the total number.</td>
    </tr>
    <tr>
        <td>flag</td>
        <td>INT</td>
        <td>-f</td>
        <td>Only output alignments with all bits in INT present in the FLAG field.</td>
    </tr>
    <tr>
        <td>lib</td>
        <td>STR</td>
        <td>-l</td>
        <td>Only output reads in library STR</td>
    </tr>
    <tr>
        <td>mapq</td>
        <td>INT</td>
        <td>-q</td>
        <td>Skip alignments with MAPQ smaller than INT</td>
    </tr>
    <tr>
        <td>readgroup</td>
        <td>STR</td>
        <td>-r</td>
        <td>Only output reads in read group STR</td>
    </tr>
</table>

<br>
License
---

Copyright (c) 2010-2012, University of Chicago
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

