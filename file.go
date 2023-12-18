package webz

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type FileUploader struct {
	FileField   string
	DstPathFunc func(*multipart.FileHeader) string
	//要考虑重名的问题
}

func (f FileUploader) Handle() HandleFunc {
	return func(ctx *Context) {
		//上传文件的逻辑在这里
		//读取文件内容
		file, fileHander, err := ctx.Req.FormFile(f.FileField)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte("上传失败")
			return
		}
		file.Close()
		//保存在哪里//将路径交给用户管理
		dst := f.DstPathFunc(fileHander)
		//可以尝试把不存在的所有目录都全部建立起来
		//os.O_WRONLY 写入数据
		//os.O_TRUNC 如果文件本身存在则清空数据
		//os.O_CREATE 不存在重建
		dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o666)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte("上传失败")
			return
		}
		defer dstFile.Close()
		//保存文件
		//最后一个参数会影响性能//可选的缓冲区 nil表示使用默认缓冲区
		io.CopyBuffer(dstFile, file, nil)
		//返回响应
		ctx.RespStatusCode = http.StatusOK
		ctx.RespData = []byte("上传成功")
	}
}

type FileDownloader struct {
	Dir string
}

func (d FileDownloader) Handle() HandleFunc {
	return func(ctx *Context) {
		req, err := ctx.QueryValue("file")
		if err != nil {
			ctx.RespStatusCode = http.StatusBadRequest
			ctx.RespData = []byte("找不到目标文件")
			return
		}
		dst := filepath.Join(d.Dir, req)
		// dst,err=filepath.Abs(dst)
		// if strings.Contains(dst,d.Dir){
		//这里可以做一个校验//防止黑客下载我们的系统文件//比如用/download?file=../../../文件名
		// }
		fn := filepath.Base(dst)
		header := ctx.Resp.Header()
		header.Set("Content-Disposition", "attachment;filename="+fn)
		header.Set("Content-Description", "File Transfer")
		header.Set("Content-Type", "application/octet-stream")
		header.Set("Content-Transfer-Encoding", "binary")
		header.Set("Expires", "0")
		header.Set("Cache-Control", "must-revalidate")
		header.Set("Pragma", "public")
		http.ServeFile(ctx.Resp, ctx.Req, dst)
	}
}

type StaticResourceHandler struct {
	dir string
}

func (s *StaticResourceHandler) Handle(ctx *Context) {
	//拿到目标文件名
	//定位到目标文件，读取出来
	//返回给前端
	file, err := ctx.PathValue("file") //这应该是拿到文件名//类似于这种 localhosy:8081/static/xxx.jpg
	if err != nil {
		ctx.RespStatusCode = http.StatusBadRequest
		ctx.RespData = []byte("请求路径参数不对")
		return
	}
	dst := filepath.Join(s.dir, file)
	data, err := os.ReadFile(dst)
	if err != nil {
		ctx.RespStatusCode = http.StatusInternalServerError
		ctx.RespData = []byte("服务器错误")
		return
	}
	ctx.RespData = data
	ctx.RespStatusCode = 200
}
