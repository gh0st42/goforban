package forban

import "net"

var MyUuid string
var MyName string
var MyPsk string

var MyFiles = []string{}
var DisableIPv6 = true
var DisableIPv4 = false
var IgnoreCount = 3

var Port = 12555

var Neighborhood map[string]ForbanNodeEntry

var HmacIgnoreList map[string]int

var DownloadQueue map[string]bool

// ServerConn server listening socket, also used for sending
var ServerConn *net.UDPConn

// BasePath Relative or absolute path where shared files and index are put
var BasePath = "var"

var FileBasePath = BasePath + "/share"

var Interfaces []string

var currentHmac string
