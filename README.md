# PING TOOL

It facilitates the connectivity check to multiple servers synchronously (with toast notifications) or asynchronously.

## USAGE:

### Async mode
ping -serverCode "id1,id2,"

### Sequential mode with (toast) and retries every 5 min ultil last serverName is available
ping -serverCode "id1,id2" --notify   

### invoke another command to run on remote server psexec.exe utility is required
ping -serverCode "id1,id2" -invokeCmd "someExecutable.exe" --notify

### invoke another command to run on remote server psexec.exe utility is required with arguments
### you can use {0} for placeholder -> will be replaces with serverId 
### or {1} placeholder -> will be replaces with server Ip Address
ping -serverCode "id1,id2" -invokeCmd "someExecutable.exe" -invokeArgs "-server {0} -option 1" --notify

### Experimental web api
ping -w 

then open a web/app or browser to localhost:4500 

### EndPoints:
* localhost:4500/sayHello
* localhost:4500/warehouses
* localhost:4500/warehouseList
