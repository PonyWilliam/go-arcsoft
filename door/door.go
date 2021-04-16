package door

import (
	serial "github.com/tarm/goserial"
	"io"
	"log"
)
var s	io.ReadWriteCloser
func FatalErr(err error){
	if err != nil{
		log.Fatal(err)
	}
}
func Send(req []byte){
	_, _ = s.Write(req)
}
func init(){
	var err error
	cfg := &serial.Config{Name: "COM9", Baud: 115200, ReadTimeout: 50}
	s,err = serial.OpenPort(cfg)
	FatalErr(err)
}