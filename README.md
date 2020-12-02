# PING TOOL

It facilitates the connectivity check to multiple servers synchronously (with toast notifications) or asynchronously.

## USAGE:

*Async mode*
ping -serverCode "101,102,"

*Sequential mode with (toast) and retries every 5 min ultil last serverName is available*
ping -serverCode "101,102" --notify   

*Experimental web api*
ping -w 

then open a web/app or browser to localhost:4500 

*EndPoints:*
* localhost:4500/sayHello
* localhost:4500/warehouses
* localhost:4500/warehouseList





