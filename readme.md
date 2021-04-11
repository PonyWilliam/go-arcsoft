# go-arcsoft
## 介绍
go-arcsoft通过移植opencv，虹软sdk，通过cgo链到c++库从而实现在golang上的人脸检测及追踪，同时封装[rfid串口](./RfidUtils/Rfid.go)在rfidutils内实现人脸追踪及上传云端
## 优点
相对于c++，golang拥有优秀的垃圾回收机制，对于内存泄漏有不可比拟的优势。在web方面有强大的库工具可以摆脱curl依赖快速进行http访问及构建http服务器，在串口方面可通过调用Windows等其它平台Api快速实现串口写入  
