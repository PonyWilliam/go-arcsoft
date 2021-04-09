package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/freetype"
	"github.com/nfnt/resize"
	"golang.org/x/image/font"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	dpi      = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "/Users/zhiliao/Downloads/ffffonts/simsun.ttf", "filename of the ttf font")
	hinting  = flag.String("hinting", "none", "none | full")
	size     = flag.Float64("size", 30, "font size in points")
	spacing  = flag.Float64("spacing", 1.5, "line spacing (e.g. 2 means double spaced)")
	wonb     = flag.Bool("whiteonblack", false, "white text on a black background")
)


var text = []string{
	"地支：沈阳市某区某镇某街道某楼某",
	"姓名：王永飞",
	"电话：1232131231232",
}

func main() {

	file, err := os.Create("dst.jpg")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	file1, err := os.Open("/Users/zhiliao/zhiliao/gopro/go_safly/src/qr.png")
	if err != nil {
		fmt.Println(err)
	}
	defer file1.Close()
	img, _ := png.Decode(file1)
	//尺寸
	img = resize.Resize(314, 314, img, resize.Lanczos3)


	jpg := image.NewRGBA(image.Rect(0, 0, 827, 1169))

	fontRender(jpg)

	draw.Draw(jpg, img.Bounds().Add(image.Pt(60, 150)), img, img.Bounds().Min, draw.Src) //截取图片的一部分
	draw.Draw(jpg, img.Bounds().Add(image.Pt(435, 150)), img, img.Bounds().Min, draw.Src) //截取图片的一部分
	draw.Draw(jpg, img.Bounds().Add(image.Pt(60, 610)), img, img.Bounds().Min, draw.Src) //截取图片的一部分
	draw.Draw(jpg, img.Bounds().Add(image.Pt(435, 610)), img, img.Bounds().Min, draw.Src) //截取图片的一部分



	png.Encode(file, jpg)

}

func fontRender(jpg *image.RGBA)  {
	flag.Parse()
	fontBytes, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		log.Println(err)
		return
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	fg, bg := image.Black, image.White
	//ruler := color.RGBA{0xdd, 0xdd, 0xdd, 0xff}
	//if *wonb {
	//	fg, bg = image.White, image.Black
	//	ruler = color.RGBA{0x22, 0x22, 0x22, 0xff}
	//}
	draw.Draw(jpg, jpg.Bounds(), bg, image.ZP, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(*dpi)
	c.SetFont(f)
	c.SetFontSize(*size)
	c.SetClip(jpg.Bounds())
	c.SetDst(jpg)
	c.SetSrc(fg)

	switch *hinting {
	default:
		c.SetHinting(font.HintingNone)
	case "full":
		c.SetHinting(font.HintingFull)
	}

	//Draw the guidelines.
	//for i := 0; i < 200; i++ {
	//	jpg.Set(10, 10+i, ruler)
	//	jpg.Set(10+i, 10, ruler)
	//}

	// Draw the text.
	pt := freetype.Pt(200, 10+int(c.PointToFixed(*size)>>6))
	for _, s := range text {
		_, err = c.DrawString(s, pt)
		if err != nil {
			log.Println(err)
			return
		}
		pt.Y += c.PointToFixed(*size * *spacing)
	}

}


func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method      //请求方法
		origin := c.Request.Header.Get("Origin")        //请求头部
		var headerKeys []string                             // 声明请求头keys
		for k, _ := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "*")        // 这是允许访问所有域
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")      //服务器支持的所有跨域请求的方法,为了避免浏览次请求的多次'预检'请求
			//  header的类型
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			//              允许跨域设置                                                                                                      可以返回其他子段
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar")      // 跨域关键设置 让浏览器可以解析
			c.Header("Access-Control-Max-Age", "172800")        // 缓存请求信息 单位为秒
			c.Header("Access-Control-Allow-Credentials", "false")       //  跨域请求是否需要带cookie信息 默认设置为true
			c.Set("content-type", "application/json")       // 设置返回格式是json
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		// 处理请求
		c.Next()        //  处理请求
	}
}
