# GoFOFA

[![Latest release](https://img.shields.io/github/v/release/FofaInfo/GoFOFA)](https://github.com/FofaInfo/GoFOFA/releases/latest)![GitHub Release Date](https://img.shields.io/github/release-date/FofaInfo/GoFOFA)![GitHub All Releases](https://img.shields.io/github/downloads/FofaInfo/GoFOFA/total)[![GitHub issues](https://img.shields.io/github/issues/FofaInfo/GoFOFA)](https://github.com/FofaInfo/GoFOFA/issues)

[英文 README](https://github.com/FofaInfo/GoFOFA/blob/main/README.md)   |  [功能使用手册]()   |  [下载](https://github.com/FofaInfo/GoFOFA/releases) 


## 项目背景

GoFOFA是一款使用Go语言编写的命令行FOFA查询工具，他除了具备基础的FOFA API接口调用能力之外，还可以直接对数据进行下一步的处理，通过模块化的调取方式，让数据从元数据到业务数据的转变。


关于GoFOFA的任何问题，欢迎加入我们的FOFA社区[微信社群](https://github.com/FofaInfo/GoFOFA/blob/74544c05a4fdd2267da35d73a7833a03f875b75e/Resource/wechat%20QRScan.jpg)或[Telegram](https://t.me/+-5xC1wYcwollYWQ1) 进行技术交流。

## 安装
在终端执行下列命令安装最新版本的GoFOFA

```
go install github.com/FofaInfo/GoFOFA/cmd/fofa@latest
```
通过配置环境变量，添加自己的FOFA API key

```
export FOFA_KEY='your_key'
```
执行测试命令`fofa search -s 1 ip=1.1.1.1`，如果返回结果，则证明安装和配置成功。

返回内容
```
fofa search -s 1 ip=1.1.1.1
query fofa of: ip=1.1.1.1
1.1.1.1,8880
```

GoFOFA拥有非常丰富的功能，查看功能使用手册和安装指南[请点击此处]()。



## 特色功能

### 批量查询和下载模块

在查询模块，除了有最基本的所有API接口的调用之外，还添加了很多实用的功能，比如批量搜索、URL拼接、证书拓线获取域名、图标多样查询和大数据量下载。

批量搜索可以解决大批量、同质化内容查询的问题，目前支持IP或者domain的批量查询。

程序可以通过调取文件中的IP，自动完成语句的拼接进行批量查询，比如查询任务为1000个IP，那么程序会自动拼接成100为1组的查询语句，共10组进行查询并自动化完成数据下载，可以大大的缩短数据调取的时间。

演示案例：

1. 1000个IP的ip.txt任务文件中；
2. 返回1年内的所有结果；
3. 获取字段包括：ip,port,host,link,title,domain;
4. 自动保存为一个data.csv格式的文件


```
fofa dump -i ip.txt -bt ip -bs 50000 -f ip,port,host,protocol,lastupdatetime -o data.csv
```

### 资产探活模块

探活，也是获取数据之后非常常见且共性的一个需求，该功能模块支持一边获取数据并同时进行探活输出，最终返回的数据会默认加上isActive字段，并会更新status_code的字段。

命令参数：`checkActive`，加上数值则代表超时情况下的重复次数。

演示案例：
1. 请求结果100条；
2. 格式选择为json格式输出；
3. 超时重复设置为3次。

```
fofa --checkActive 3 -s 100 --format=json port=80 && type=subdomain
```
### JS渲染模块

JS渲染模块顾名思义可以针对获取的数据中的URL字段进行JS渲染处理，处理后支持选择获取渲染后的html标签，目前支持同步获取渲染后的title、body字段。

命令参数：`jsRender` 选择渲染模块；`-url`，用来选择单个目标；`-tags`选择获取渲染后的标签。

:red_circle: 注意：jsRender模块对性能要求过高，不建议随意更改workers参数。

演示案例：

1. 请求结果10条；
2. 从管道中获取到的url进行渲染识别；
3. 并标记上新获取的title字段。

```
fofa jsRender -tags title --format=json -i link.txt -o data.txt
```

### 资产分类模块

在结果输出环节，GoFOFA支持通过关键特征对CSV文件进行分类。这一操作可以通过config.yaml配置文件来完成。

我们可以提前在config.yaml文件中设置过滤规则filter，通过内置一个contain的方法对数据进行处理并分类。

调取方式：

```
fofa category -input input.csv [-o category.csv]
```

规则配置案例：

1. sheet1：资产标题中包含关键词"后台”、”系统”、”登录”、“管理”、“门户” 字样的网站；
2. sheet2：资产标题中包含关键词“nginx”、“tomact”、“IIS”、“Welcome to OpenResty”字样的资产；
3. sheet3：非http/https协议的资产；
4. sheet4: category字段包含以下分类“路由器”、“视频监控”、“网络存储”、“防火墙”标签的资产；

```
categories:  
	- name: "sheet1"    
	  filters:    
		- "CONTAIN(title, '后台') && CONTAIN(title, '登录')"
		- "CONTAIN(title, '管理') && CONTAIN(title, '系统')"

  	- name: "sheet2"    
	  filters:      
		- "CONTAIN(title, 'nginx') && CONTAIN(title, 'tomcat')"
 		- "CONTAIN(title, 'IIS') && CONTAIN(title, 'Welcome to OpenRestry')"

	- name: "sheet3"
	  filters:
		- "protocol!='http' || protocol!=''https"

	- name: "sheet4"
	  filters:
		- "CONTAIN(product_category, '路由器') && CONTAIN(product_category, '视频监控')"
 		- "CONTAIN(product_category, '网络存储') && CONTAIN(product_category, '防火墙')"

```



## GoFOFA所有参数示例 

### search

| 参数        | 参数简写 | 默认值  | 简介                                              |
| ----------- | -------- | ------- | ------------------------------------------------- |
| fields      | f        | ip,port | FOFA返回的字段选择，[了解更多](https://fofa.info/vip) |                             
| format      |          | csv     | 输出格式，可以为csv/json/xml                      |
| outFile     | o        |         | 输出文件，如果不设置则终端打印                    |
| size        | s        | 100     | 查询数量，最大为10000，受deductMode参数限制       |
| deductMode  |          |         | 消费f点数，不设置则读取用户最大免费数量           |
| fixUrl      |          | false   | 是否组合url，例如1.1.1.1,80组合为http://1.1.1.1   |
| urlPrefix   |          | http:// | url前缀                                           |
| full        |          | false   | 是否调取全量数据                                  |
| uniqByIP    |          | false   | 是否根据ip去重                                    |
| workers     |          | 10      | 线程数量                                          |
| rate        |          | 2       | 每秒查询次数                                      |
| template    |          | ip={}   | 从管道获取输入，输入的内容会替换{}                |
| inFile      | i        |         | 输入文件，如果不设置则读取管道输入                |
| checkActive |          | -1      | 探活复测次数，-1为不使用探活                      |
| deWildcard  |          | -1      | 泛解析去重，-1为不使用泛解析去重                  |
| filter      |          |         | 数据过滤规则，例如port<100 || host=="baidu.com" |
| dedupHost   |          | false   | subdomain去重                                     |
| headline    |          | false   | 是否输出csv头，只有在format为csv时可用            |
| help        | h        | false   | 使用方法                                          |

### dump

| 参数      | 参数简写 | 默认值  | 简介                                                  |
| --------- | -------- | ------- | ----------------------------------------------------- |
| fields    | f        | ip,port | FOFA返回的字段选择，[了解更多](https://fofa.info/vip) |
| format    |          | csv     | 输出格式，可以为csv/json/xml                          |
| outFile   | o        |         | 输出文件，如果不设置则终端打印                        |
| inFile    | i        |         | 输入文件，如果不设置则读取管道输入                    |
| size      | s        | 100     | 查询数量，无上限，但要扣除f点或免费数量               |
| fixUrl    |          | false   | 是否组合url，例如1.1.1.1,80组合为http://1.1.1.1       |
| urlPrefix |          | http:// | url前缀                                               |
| full      |          | false   | 是否调取全量数据                                      |
| batchSize | bs       | 1000    | 每次拉取多少条数据                                    |
| batchType | bt       |         | 批量查询，可以为ip/domain                             |
| help      | h        | false   | 使用方法                                              |

### jsRender

| 参数    | 参数简写 | 默认值 | 简介                               |
| ------- | -------- | ------ | ---------------------------------- |
| url     | u        |        | 单个url渲染                        |
| tags    | t        |        | 获取标签，目前可以为title/body     |
| format  |          | csv    | 输出格式，可以为csv/json/xml       |
| outFile | o        |        | 输出文件，如果不设置则终端打印     |
| inFile  | i        |        | 输入文件，如果不设置则读取管道输入 |
| workers |          | 2      | 线程数量                           |
| retry   |          | 3      | 超时尝试次数                       |
| help    | h        | false  | 使用方法                           |

### domains

| 参数       | 参数简写 | 默认值 | 简介                                        |
| ---------- | -------- | ------ | ------------------------------------------- |
| outFile    | o        |        | 输出文件，如果不设置则终端打印              |
| size       | s        | 100    | 查询数量，最大为10000，受deductMode参数限制 |
| deductMode |          |        | 消费f点数，不设置则读取用户最大免费数量     |
| full       |          | false  | 是否调取全量数据                            |
| withCount  |          | false  | 是否输出域名数量                            |
| clue       |          | false  | 是否输出线索语句                            |
| help       | h        | false  | 使用方法                                    |

### active

| 参数    | 参数简写 | 默认值 | 简介                               |
| ------- | -------- | ------ | ---------------------------------- |
| url     | u        |        | 单个url存活探测                    |
| format  |          | csv    | 输出格式，可以为csv/json/xml       |
| outFile | o        |        | 输出文件，如果不设置则终端打印     |
| inFile  | i        |        | 输入文件，如果不设置则读取管道输入 |
| workers |          | 2      | 线程数量                           |
| retry   |          | 3      | 超时尝试次数                       |
| help    | h        | false  | 使用方法                           |

### category

| 参数   | 参数简写 | 默认值 | 简介                    |
| ------ | -------- | ------ | ----------------------- |
| inFile | i        |        | 输入分类文件，可以为csv |
| unique |          |        | 分类数据是否唯一性      |
| help   | h        | false  | 使用方法                |

### dedup

| 参数    | 参数简写 | 默认值        | 简介                          |
| ------- | -------- | ------------- | ----------------------------- |
| dedup   | d        |               | 需要去重的字段                |
| inFile  | i        |               | 输入需要去重的文件，可以为csv |
| outFile | o        | duplicate.csv | 输出文件                      |
| help    | h        | false         | 使用方法                      |

### host

| 参数 | 参数简写 | 默认值 | 简介     |
| ---- | -------- | ------ | -------- |
| help | h        | false  | 使用方法 |

### icon

| 参数 | 参数简写 | 默认值 | 简介                                 |
| ---- | -------- | ------ | ------------------------------------ |
| open |          | false  | 是否根据icon计算结果打开fofa搜索页面 |
| help | h        | false  | 使用方法                             |

### stats

| 参数   | 参数简写 | 默认值        | 简介                                                  |
| ------ | -------- | ------------- | ----------------------------------------------------- |
| fields | f        | title,country | FOFA返回的字段选择，[了解更多](https://fofa.info/vip) |
| size   | s        | 5             | 查询次数，-1表示永远不停                              |
| help   | h        | false         | 使用方法                                              |

### random

| 参数      | 参数简写 | 默认值                                          | 简介                                                  |
| --------- | -------- | ----------------------------------------------- | ----------------------------------------------------- |
| fields    | f        | ip,port,host,header,title,server,lastupdatetime | FOFA返回的字段选择，[了解更多](https://fofa.info/vip) |
| format    |          | json                                            | 输出格式，可以为csv/json/xml                          |
| size      | s        | 1                                               | 查询次数，-1表示永远不停                              |
| sleep     |          | 1000                                            | 获取间隔，单位ms                                      |
| fixUrl    |          | false                                           | 是否组合url，例如1.1.1.1,80组合为http://1.1.1.1       |
| urlPrefix |          | http://                                         | url前缀                                               |
| full      |          | false                                           | 是否调取全量数据                                      |
| help      | h        | false                                           | 使用方法                                              |

### count

| 参数 | 参数简写 | 默认值 | 简介     |
| ---- | -------- | ------ | -------- |
| help | h        | false  | 使用方法 |

### account

| 参数 | 参数简写 | 默认值 | 简介     |
| ---- | -------- | ------ | -------- |
| help | h        | false  | 使用方法 |



## 最后的碎碎念

在决定开源 GoFOFA 之前，我们深知社区中已经有许多优秀的 FOFA API 调取工具。比如由 WgpSec 狼组的 [fofa_viewer](https://github.com/wgpsec/fofa_viewer)、Xiecat的[fofax](https://github.com/xiecat/fofax)、以及HxO 战队的[FofaMap](https://github.com/asaotomo/FofaMap)等。这些工具功能强大，且深受白帽子们的喜爱。起初我们也曾认为，既然已经有这么多出色的 FOFA API 工具，似乎作为官方再开发一个类似的工具显得多余。

然而，随着我们对用户需求的深入了解，发现仅仅停留在 API 数据调取远远不够。在将数据拉取下来后，如何进行高效的二次处理，从而满足各种实际场景需求，是一个巨大的痛点。我们深知，“只有想不到，没有师傅们做不到”，各种数据处理需求层出不穷，包括我们内部在进行数据处理过程中，也实际在做各种各样的开发。

因此，我们萌生了一个想法：为什么不整合共性需求开源出来一款工具，既能高效调取数据，又能直接提供数据的二次处理能力，同时成为一个社区共创的平台？用户不仅可以体验 FOFA 的功能，还可以提出更多需求，甚至与我们一同共创未来的工具功能。

我们希望通过开源 GoFOFA，让更多用户体验这款功能强大且灵活的工具，同时期待与师傅们一起将 FOFA 社区的技术创新与共创共同发展。

如果你觉得这个工具还不错的话，给我们点个Star吧～

