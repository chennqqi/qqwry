# qqwry

IP纯真库 golang 解析

forked from github.com/freshcn/qqwry

对原版API做了一下格式调整和API封装
原版服务移动到examples目录下

demo
```
	err := qqwry.Init("qqwry.dat")
	if err != nil {
		log.Panic(err)
	}
	count, _ := qqwry.Count()
	du := time.Since(startTime)
	log.Println("IP load:", count, " records, cost:", du)
	
	qqWry, _ := qqwry.NewQQwry()
	if res, err:=qqWry.Query("10.10.10.10");err!=nil{
		log.Println(err)
	} else {
		log.Println(res)
	}


```

### go安装

```
go get -u github.com/chennqqi/qqwry
```

### 下载纯真IP库
请访问 http://www.cz88.net 下载纯真IP库，需要在windows中安装程序，然后在程序的安装目录可以找到qqwry.dat文件，复制出来放到和本程序同一个目录（当然也可是其他目录，只是需要在运行的时候指定IP库目录）

## 使用方法

* ip - 输入的ip地址
* country - 国家或地区
* area - 区域（我实际测试得到还有可能是运营商）


### 感谢

* 感谢[纯真IP库](http://www.cz88.net)一直以来坚持为大家提供免费的IP库资源
