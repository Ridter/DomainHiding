# DomainHiding External C2 Demo

## Domain Hiding

https://github.com/SixGenInc/Noctilucent

![](https://blogpics-1251691280.file.myqcloud.com/imgs/20201027105320.png)

利用此技术可实现对C2的隐藏，但是国内这个技术不能使用，为啥？可以看[这里](https://geneva.cs.umd.edu/posts/china-censors-esni/esni/)。


## This Demo
这个项目是一个demo，代码写的有点戳，就是简要实现了一下利用此技术的External C2。

需要准备的东西：
```
1、使用cloudflare解析的域名。
2、一个VPS（C2），另外DNS上需要设置将域名指向此VPS
```

寻找用来伪装的域名。可以看[这里](https://github.com/SixGenInc/Noctilucent/tree/master/findfronts)。

替换`main.go` 中相关项
```
frontDomain     替换要伪造成的域名
actualDomain    替换成自己的域名
pipeName        自定义一个管道名称
```

Cobaltstrike开启External c2
```
externalc2_start("0.0.0.0",2222);
```

在VPS上运行`server.py`

```
sudo python server.py
```

运行client
```
go run main.go
```

最终效果：

![](https://blogpics-1251691280.file.myqcloud.com/imgs/20201027112014.png)

DNS数据：
![](https://blogpics-1251691280.file.myqcloud.com/imgs/20201027112152.png)

查询VPS ip：

![](https://blogpics-1251691280.file.myqcloud.com/imgs/20201027112245.png)


非常感谢以下优秀的项目作为本项目支撑：
```
https://github.com/Lz1y/GECC
https://github.com/SixGenInc/Noctilucent
https://github.com/mdsecactivebreach/Browser-ExternalC2
```