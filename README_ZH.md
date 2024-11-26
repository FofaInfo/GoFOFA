# GoFOFA

[![Latest release](https://img.shields.io/github/v/release/FofaInfo/GoFOFA)](https://github.com/FofaInfo/GoFOFA/releases/latest)![GitHub Release Date](https://img.shields.io/github/release-date/FofaInfo/GoFOFA)![GitHub All Releases](https://img.shields.io/github/downloads/FofaInfo/GoFOFA/total)[![GitHub issues](https://img.shields.io/github/issues/FofaInfo/GoFOFA)](https://github.com/FofaInfo/GoFOFA/issues)

[:blue_book: 英文 README](https://github.com/FofaInfo/GoFOFA/blob/main/README.md)   |   [:floppy_disk:  下载](https://github.com/FofaInfo/GoFOFA/releases/tag/v0.2.25)   |   [:orange_book:  FOFA API文档](https://fofa.info/api)


## 项目背景

GoFOFA是一款使用Go语言编写的命令行FOFA查询工具，他除了具备基础的FOFA API接口调用能力之外，还可以直接对数据进行下一步的处理，通过模块化的调取方式，让数据从元数据到业务数据的转变。

我们在里面融合了很多FOFA工程师常用的调取小功能和数据处理的功能，如果您有更多的想法和需求，欢迎随时在issue中进行提交。


关于GoFOFA的任何问题，欢迎加入我们的FOFA社区[微信社群](https://github.com/FofaInfo/GoFOFA/blob/74544c05a4fdd2267da35d73a7833a03f875b75e/Resource/wechat%20QRScan.jpg)或[Telegram](https://t.me/+-5xC1wYcwollYWQ1) 进行技术交流。

## 目录
### 配置
### 数据查询模块

- [基础查询](#基础查询)

- [查询实用功能](#查询实用功能)
	- [批量搜索（支持txt上传进行批量查询）](#批量搜索)
	- [指定URL拼接](#URL拼接)
	- [随机从FOFA生成数据](#随机从FOFA生成数据)
	- [证书拓线获取域名](#证书拓线查询域名)
	- [icon多样查询](#Favicon图标查询)
	- [大数据量下载](#大数据量下载)
- [统计聚合接口](#统计聚合接口)
- [HOST聚合接口](#HOST聚合接口)


### 数据处理模块

- [URL去重](#URL去重)
- [泛解析去重](#泛解析去重)
- [存活探测（支持从管道批量输入）](#存活探测（支持从管道批量输入)
- [JS渲染识别（支持从管道批量输入）](#JS渲染识别（支持从管道批量输入）)
- [数据资产分类](#数据资产分类)

注意：部分数据处理功能模块有必要的字段要求，清在数据调取时确认包含该字段。

### 其他

- [Gofa版本号](#其他)
- [GoFOFA所有参数示例](#GoFOFA所有参数示例)

## 配置

- 下载gofofa:

```
$ go install github.com/FofaInfo/GoFOFA/cmd/fofa@latest
```

- 显示如下表示安装成功:

```
$ fofa
NAME:
   fofa - fofa client on Go v0.2.25, commit none, built at unknown

USAGE:
   fofa [global options] command [command options] [arguments...]

VERSION:
   v0.2.25

AUTHOR:
   LubyRuffy <lubyruffy@gmail.com>
   Y13ze <y13ze@outlook.com>

COMMANDS:
   search   fofa host search
   account  fofa account information
   count    fofa query results count
   stats    fofa stats
   icon     fofa icon search
   random   fofa random data generator
   host     fofa host
   dump     fofa dump data
   domains  extend domains from a domain
   active   website active
   dedup    remove duplicate tool
   category  classify data according to config
   jsRender  website js render
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --fofaURL value, -u value  format: <url>/?key=<key>&version=<v2> (default: "https://fofa.info/?key=your_key&version=v1")
   --verbose                  print more information (default: false)
   --accountDebug             print account in error log (default: false)
   --help, -h                 show help (default: false)
   --version, -v              print the version (default: false)
```

- 配置环境变量:

```
$ export FOFA_CLIENT_URL='https://fofa.info/?key=your_key'
```
或者：
```
$ export FOFA_KEY='your_key'
```
注意：FOFA_CLIENT_URL的优先级最高

#### MacOS/Linux版本

可以将下载的 GoFOFA 压缩包解压放在 `/usr/local/bin/` 目录下，这样的好处是在终端任何一个位置都可以使用这个命令。

```text
tar -zxvf ~/Downloads/fofa_0.2.25_darwin_amd64.tar.gz -C /usr/local/bin/
```

#### Windows版本

解压压缩包，直接运行fofa.exe。

## 功能介绍

### 数据查询模块

#### 基础查询

fofa语法查询，可以输入单个查询语句，默认返回字段是ip和端口。

```shell
$ fofa search port=80
2024/08/23 11:51:19 query fofa of: port=80
69.10.146.92,80
20.193.138.22,80
194.182.72.64,80
......
......
```

也可以输入语法组合式进行查询。

```shell
$ fofa search 'port=80 && protocol=ftp'
2024/08/23 11:52:00 query fofa of: port=80 && protocol=ftp
139.196.102.155,80
59.82.133.71,80
69.80.101.32,80
69.80.101.68,80
......
......
```

不选择子模块直接查询的话，会默认使用search模块进行查询:

```shell
$ fofa domain=qq.com
2024/08/23 11:53:00 query fofa of: domain=qq.com
14.22.33.13,443
183.47.126.116,443
14.22.33.13,443
14.22.33.13,443
......
......
```

使用`--fields`来选择输出的字段，会根据选定的字段进行返回，下面的示例选择了`host,ip,port,protocol,lastupdatetime`字段。

```shell
$ fofa search --fields host,ip,port,protocol,lastupdatetime 'port=6379'
2024/08/23 12:09:08 query fofa of: port=6379
168.119.197.62:6379,168.119.197.62,6379,redis,2024-08-23 12:00:00
119.45.170.222:6379,119.45.170.222,6379,redis,2024-08-23 12:00:00
112.126.87.29:6379,112.126.87.29,6379,unknown,2024-08-23 12:00:00
121.43.116.245:6379,121.43.116.245,6379,unknown,2024-08-23 12:00:00
......
......
```

或者更简洁一些，直接使用`-f`来表示字段。

```shell
$ fofa search -f host,ip,port,protocol,lastupdatetime 'port=6379'
2024/08/23 12:09:08 query fofa of: port=6379
168.119.197.62:6379,168.119.197.62,6379,redis,2024-08-23 12:00:00
119.45.170.222:6379,119.45.170.222,6379,redis,2024-08-23 12:00:00
112.126.87.29:6379,112.126.87.29,6379,unknown,2024-08-23 12:00:00
121.43.116.245:6379,121.43.116.245,6379,unknown,2024-08-23 12:00:00
......
......
```



使用`--size`来选择单词输出的数据返回量， 默认大小是100:

```shell
$ fofa search --size 5 'port=6379'
2024/08/23 14:07:18 query fofa of: port=6379
47.99.89.216,6379
112.124.14.11,6379
107.154.224.11,6379
39.101.36.243,6379
139.196.136.107,6379
```

或者更简洁一些，直接使用`-s`来表示字段。

```shell
$ fofa search -s 5 'port=6379'
2024/08/23 14:07:18 query fofa of: port=6379
47.99.89.216,6379
112.124.14.11,6379
107.154.224.11,6379
39.101.36.243,6379
139.196.136.107,6379
```

如果需要输出不同的数据格式，可以通过`--format`来设置，默认是`csv`格式，还支持`json`和`xml`格式:

```shell
$ fofa search --format=json 'port=6379'
2024/08/23 14:05:49 query fofa of: port=6379
{"ip":"39.101.36.243","port":"6379"}
{"ip":"139.196.136.107","port":"6379"}
{"ip":"47.97.53.84","port":"6379"}
{"ip":"39.104.71.245","port":"6379"}
......
......
$ fofa search --format=xml 'port=6379'
2024/08/23 14:08:19 query fofa of: port=6379
<result><port>6379</port><ip>39.101.36.96</ip></result>
<result><ip>47.99.89.216</ip><port>6379</port></result>
<result><ip>112.124.14.11</ip><port>6379</port></result>
<result><ip>23.224.60.162</ip><port>6379</port></result>
......
......
```

或者通知使用`--format json`或者`--format xml`直接进行更换。

```shell
$ fofa search --format json 'port=6379'
2024/08/23 14:05:49 query fofa of: port=6379
{"ip":"39.101.36.243","port":"6379"}
{"ip":"139.196.136.107","port":"6379"}
{"ip":"47.97.53.84","port":"6379"}
{"ip":"39.104.71.245","port":"6379"}
......
......
$ fofa search --format xml 'port=6379'
2024/08/23 14:08:19 query fofa of: port=6379
<result><port>6379</port><ip>39.101.36.96</ip></result>
<result><ip>47.99.89.216</ip><port>6379</port></result>
<result><ip>112.124.14.11</ip><port>6379</port></result>
<result><ip>23.224.60.162</ip><port>6379</port></result>
......
......
```

使用`--outFile`可以将结果输出到指定文件中，若不设置次参数则默认输出在命令行中:

```shell
$ fofa search --outFile a.txt 'port=6379'
```

或者直接使用`-o`指定输出到指定文件中。

```shell
$ fofa search -o a.txt 'port=6379'
```

账号信息查询接口，通过`account`可以获取账户信息。

```shell
$ fofa account
{
  "error": false,
  "fcoin": 0,
  "fofa_point": 99982,
  "isvip": true,
  "vip_level": 5,
  "remain_api_query": 4999635,
  "remain_api_data": 49949766
}
```

FOFA查询结果数量统计，可以通过`count`模块统计数据数量。

```shell
$ fofa count port=80
587055296
```

### 查询实用功能
#### 批量搜索

通过dump模块的--batchType参数开启批量搜索，目前可以生成批量搜索的字段包括：ip,domain。

ip批量搜索会将文件中的ip以100为一组生成批量搜索的语句，domain批量搜索会将文件中的domain以50为一组生成批量搜索的语句。

如文件中有1000个ip，将自动生成10组查询语句完成批量搜索。


```shell
$ cat ip.txt
106.75.95.206
123.58.224.8
$ fofa dump -i ip.txt --batchType ip
2024/11/25 14:51:10 dump data of query: ip=106.75.95.206 || ip=123.58.224.8
123.58.224.8,40544
123.58.224.8,31497
106.75.95.206,80
......
```

或者更简洁一些:

```shell
$ fofa dump -i ip.txt -bt ip
2024/11/25 14:51:10 dump data of query: ip=106.75.95.206 || ip=123.58.224.8
123.58.224.8,40544
123.58.224.8,31497
106.75.95.206,80
......
```

一般都会输出到文件中，输出结果会打印输出进度:

```shell
$ fofa dump -i ip.txt -bt ip -o dump.csv
2024/11/25 14:51:10 dump data of query: ip=106.75.95.206 || ip=123.58.224.8
2024/11/25 14:52:03 size: 188/188, 100.00%
......
```

--batchSize可以用来设置每次拉取的数量，默认为1000。如一组批量查询结果有20000条数据，默认每次拉取数量为1000条，则需要执行拉取20次，完成后继续执行下一组查询语句。

--size为每组需要拉取的总数据量，默认为-1代表获取所有数据。

```shell
$ fofa dump -i ip.txt -bt ip -o dump.csv
2024/11/25 17:39:05 dump data of query: ip=112.25.151.122 || ... || ip=58.213.160.221
2024/11/25 17:39:06 size: 115/115, 100.00%
2024/11/25 17:39:06 dump data of query: ip=221.226.119.3 || ... || ip=221.226.6.2
2024/11/25 17:39:12 size: 153/153, 100.00%
```

#### URL拼接

1. 如果你想获取完整的url拼接，可以使用`fixUrl`参数:

```shell
$ fofa --size 2 --fields "host" title=Gitblit
2024/08/23 14:23:02 query fofa of: title=Gitblit
pmsningbo.veritrans.cn:20202
platform.starpost.cn:8080

$ fofa --size 2 --fields "host" --fixUrl title=Gitblit
2024/08/23 14:23:34 query fofa of: title=Gitblit
http://pmsningbo.veritrans.cn:20202
http://platform.starpost.cn:8080
```
2. 如果你想要使用其他的前缀，可以使用`urlPrefix`参数设置前缀:

```shell
$ fofa --size 1 --fields "host" --fixUrl --urlPrefix "redis://" protocol=redis
2024/08/23 14:29:26 query fofa of: protocol=redis
redis://139.9.222.14:7000
```

#### 随机从FOFA生成数据


通过`random`命令随机从fofa生成数据:

```shell
$ fofa random -f host,ip,port,lastupdatetime,title,header,body --format json
{"body":"","header":"HTTP/1.1 401 Unauthorized\r\nWww-Authenticate: Digest realm=\"IgdAuthentication\", domain=\"/\", nonce=\"ZjVhNGY2YzI6MTUyNDM2N2Y6MzRiMGZjZjQ=\", qop=\"auth\", algorithm=MD5\r\nContent-Length: 0\r\n","host":"95.22.200.127:7547","ip":"95.22.200.127","lastupdatetime":"2024-08-14 13:00:00","port":"7547","title":""}
```

可以通过sleep参数设置时间500ms，按照时间每500ms生成一次数据:

```shell
$ fofa random -s -1 -sleep 500
```

#### 证书拓线查询域名

domains子模块主要用于最简单的拓线，通过证书进行拓线，可以使用`withCount`来统计数量，来获取更多的数据。

该功能必要获取的FOFA字段：`certs_domains、certs_subject_org`

```shell
$ fofa domains -s 1000 -withCount baidu.com
baidu.com       660
dwz.cn  620
dlnel.com       614
bcehost.com     614
bdstatic.com    614
......
......
```

你还可以使用 `-uniqByIP` 来去除相同的ip:
```shell
$ fofa domains -s 1000 -withCount -uniqByIP baidu.com 
baidu.com       448
dwz.cn  410
aipage.cn       406

```

#### Favicon图标查询

Favicon图标查询及Hash值计算

直接查询的三种方式：
1. 你可以通过读取本地的ico文件来查询数据，open参数会自动帮你跳转到fofa:

```shell
$ fofa icon --open ./data/favicon.ico
```

2.也可以通过网页的ico文件来查询:

```shell
$ fofa icon --open https://fofa.info/favicon.ico
```

3. 还可以直接通过url来查询:

```shell
$ fofa icon --open http://www.baidu.com
```

计算图标Hash值的三种方式：

1. 获取本地ico文件的hash值:

```shell
$ fofa icon ./data/favicon.ico
-247388890
```

2. 也可以获取网页ico文件的hash值:

```shell
$ fofa icon https://fofa.info/favicon.ico
-247388890
```

3. 还可以直接获取url的ico_hash值:

```shell
$ fofa icon http://www.baidu.com
-1588080585
```

#### 大数据量下载

大批量数据下载使用，使用`--batchSize`设置每组下载数量，一键完成数据下载并存储到指定文件:

```shell
$ fofa dump --format json -fixUrl -outFile a.json -batchSize 500 'title=phpinfo'
```

通过fofa语句文件，来下载并存储大数据（每条数据一行）:

```shell
cat queries.txt
port=13344
port=23455

# csv
$ fofa dump -outFile out.csv -inFile queries.txt

# json
$ fofa dump -inFile queries.txt -outFile out.json -j
2023/08/09 10:05:33 dump data of query: port=13344
2023/08/09 10:05:35 size: 11/11, 100.00%
2023/08/09 10:05:35 dump data of query: port=23455
2023/08/09 10:05:37 size: 499/499, 100.00%
```

### 统计聚合接口

数据统计接口调取，stats模块可以做数据统计等操作

```shell
$ fofa stats --fields title,country title="hacked by"
===  title
Hacked By Ashiyane Digital Security Team        706
Hacked By MR.GREEN      465
Hacked by Kn1gh7        259
Hacked By MR.GREEN &#8211; Just another WordPress site  163
HackeD By Desert Warriors       108
===  country
美国    3182
德国    259
波兰    225
英国    223
新加坡  205
```
### HOST聚合接口

HOST聚合接口调取，通过Host模块，输入域名即可获取host视角的资产信息:

```shell
$ fofa host demo.cpanel.net
Host:            demo.cpanel.net
IP:              208.74.120.133
ASN:             33522
ORG:             CPANEL-INC
Country:         United States of America
CountryCode:     US
Ports:           [2078 3306 2079 2082 143 993 2086 2095 2083 2087 110 2080 80 995 2096 2077 443]
Protocols:       imaps,mysql,https,imap,pop3s,http,pop3
Categories:      Server Management
Products:        cPanel-MGMT-Products
UpdateTime:      2022-05-30 17:00:00
```

### 数据处理模块

#### URL去重

数据去重可以通过两种方式进行。

1. 第一种是边查询边进行去重操作，即从FOFA拉下来的数据就是去重完成的。```--dedupHost```，在fofa中subdomain代表网页数据，service代表协议数据，如果host相同，优先保留subdomain数据:

```shell
$ fofa search -s 3 -f host,type "ip=106.75.95.206"
2024/08/28 19:49:23 query fofa of: ip=106.75.95.206
106.75.95.206,subdomain
106.75.95.206:443,service
106.75.95.206,service
$ fofa search -s 3 -f host,type --dedupHost "ip=106.75.95.206"
2024/08/28 19:52:30 query fofa of: ip=106.75.95.206
https://106.75.95.206,subdomain
106.75.95.206:443,service
106.75.95.206,subdomain
```

2. 第二种支持文件上传的形式对已有数据的任意字段进行去重操作。`dedup`命令支持对一个csv文件中的某一个字段进行去重，通过input参数上传文件，通过dedup参数选择去重字段（会根据字段顺序进行去重），通过output设置输出文件名（默认duplicate.csv）:

```shell
$ fofa dedup -output data.csv -dedup ip -output dedup.csv
$ fofa dedup -output data.csv -dedup ip,host,domain -output dedup.csv
```
或者可以更简洁一些:

```shell
$ fofa dedup -o data.csv -d ip -o dedup.csv
$ fofa dedup -o data.csv -d ip,host,domain -o dedup.csv
```


#### 泛解析去重

所需获取的FOFA字段：ip、port、domain、title、fid

如果你想要减少泛域名数量，可以使用```--deWildcard```设置保留泛域名数量，```-f```可以支持其他字段选用link做为演示:

```shell
$ fofa search -s 3 -f link domain=huashunxinan.net
2024/08/27 17:19:04 query fofa of: domain=huashunxinan.net
http://h8huumr2zdmwgy5.huashunxinan.net
http://keygatjexlvsznh.huashunxinan.net
http://jobs.huashunxinan.net
$ fofa search -s 3 -f link --deWildcard 1 domain=huashunxinan.net
2024/08/27 17:26:42 query fofa of: domain=huashunxinan.net
http://h8huumr2zdmwgy5.huashunxinan.net
https://fwtn2k7oigaiyla.huashunxinan.net
http://huashunxinan.net
```

#### 存活探测（支持从管道批量输入）

存活探测可以通过两种方式进行。

1. 第一种模式必要获取字段：link，实现方式是判断用户输入的是否有需要字段，没有则添加上，最终返回的数据只剩下用户输入的字段并在最后加上isActive字段。

可以使用```--checkActive 3```，`3`是超时重复次数（使用这个参数之后也会重新获取status_code数据）:

```shell
$ fofa search -s 3 --checkActive port=80 
2024/08/26 18:52:00 query fofa of: port=80
216.92.244.44,80,true
104.21.31.50,80,true
182.247.239.68,80,true
$ fofa search -s 3 --checkActive 3 --format=json port=80
2024/08/26 18:53:33 query fofa of: port=80
{"ip":"54.78.179.223","isActive":"false","port":"80"}
{"ip":"18.155.202.65","isActive":"true","port":"80"}
{"ip":"198.144.179.122","isActive":"true","port":"80"}
$ fofa search -s 3 --checkActive 3 --format=xml port=80
2024/08/26 18:54:38 query fofa of: port=80
<result><ip>50.16.35.210</ip><port>80</port><isActive>true</isActive></result>
<result><ip>104.21.77.62</ip><port>80</port><isActive>true</isActive></result>
<result><ip>189.193.236.170</ip><port>80</port><isActive>false</isActive></result>
```

2. 第二种模式支持从管道输入或者从文件输入url，可以使用`--url`来获取存活信息，true为存活，false为不存活:

```shell
$ fofa active --url baidu.com,fofa.info,asdsadsasdas.com
baidu.com,true
fofa.info,true
asdsadsasdas.com,false
```

或者可以更简洁一些:

```shell
$ fofa active -u baidu.com,fofa.info,asdsadsasdas.com
baidu.com,true
fofa.info,true
asdsadsasdas.com,false
```

还可以通过对一个每行为一个url的文件进行探测:

```shell
$ cat target.txt
baidu.com
fofa.info
asdsadsasdas.com
$ fofa active -i target.txt  
baidu.com,true
fofa.info,true
asdsadsasdas.com,false
```

还支持对管道中的url进行探测（管道中的数据需为每行一条url）:

```shell
$ fofa search -f link -s 3 port=80 | fofa active
2024/08/23 15:50:11 query fofa of: port=80
http://og823.hb-yj.com,true
http://rw823.tcxzgh.org,true
http://sb823.tcxzgh.org,true
```

#### JS渲染识别（支持从管道批量输入）

必要获取的FOFA字段：link、需要完成search任务之后，进行单独的渲染识别

jsRender模块用来对url进行js渲染，支持选择获取渲染后的html标签，目前支持获取标签title、body，可以使用`-url`来选择单个目标，`-tag`选择获取渲染后的标签：

```shell
$ fofa jsRender -url http://baidu.com -tag title
http://baidu.com,百度一下，你就知道
```

或者可以更简洁一些:

```shell
$ fofa jsRender -u http://baidu.com -t title
http://baidu.com,百度一下，你就知道
```

还可以通过对一个每行为一个url的文件进行探测:

```shell
$ cat url.txt
http://baidu.com
https://fofa.info
$ fofa jsRender -i url.txt -t title 
http://baidu.com,百度一下，你就知道
https://fofa.info,网络空间测绘，网络空间安全搜索引擎，网络空间搜索引擎，安全态势感知 - FOFA网络空间测绘系统
```

还支持对管道中的url进行探测（管道中的数据需为每行一条url）:

```shell
$ fofa search -f link -s 3 port=80 | fofa jsRender -t title
2024/08/23 15:50:11 query fofa of: port=80
http://project5.abioyibo.com,Just another WordPress site
http://www.valuegoodsbazaar.shop,srv258.sellvir.com — Coming Soon
http://forecasting-preprod.pcasys.co.uk,- Sales Forecasting Tool (preprod)
```

#### 数据资产分类

Category支持对一个csv文件进行分类，通过config.yaml配置文件来进行分类（配置文件必须在当前目录下），配置文件如下格式:

```shell
categories:
  - name: "百度贴吧"
    filters:
      - "(protocol == 'http' || protocol == 'http') && CONTAIN(title, '百度贴吧')"
      - "domain=='baidu.com' && CONTAIN(title, '百度贴吧')"

  - name: "百度3xx页面"
    filters:
      - "domain=='baidu.com' && status_code >= '300' && status_code < '400'"

  - name: "其他"
    filters:
      - "CONTAIN(title, '百度')"

```

可以在config.yaml文件中设置好过滤规则`filter`，内置了一个`CONTAIN`方法，意思是某一个字段是否含有什么值`-output`不设置会默认生成`category.csv`文件:

```shell
$ fofa category -input input.csv [-output category.csv]
```

或者更简洁一些:

```shell
$ fofa category -i input.csv [-o category.csv]
```

### 其他

获取gofofa版本号

```shell
$ fofa --version
```


#### GoFOFA所有参数示例 

| 参数        | 参数简写 | 默认值  | 简介                                              |
| ----------- | -------- | ------- | ------------------------------------------------- |
| fields      | f        | ip,port | 获取fofa字段                                      |
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
| filter      |          |         | 数据过滤规则，例如`port<100 || host=="baidu.com"` |
| dedupHost   |          | false   | subdomain去重                                     |
| headline    |          | false   | 是否输出csv头，只有在format为csv时可用            |
| help        | h        | false   | 使用方法                                          |
