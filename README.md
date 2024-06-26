## 简介
> [!NOTE]
> 此项目处于快速迭代期，会出现破坏性更新，每次更新程序时，需要查看release的说明。谨慎更新。

我家猫看B站发现有些收藏的视频失效了，于是他自己用GO写了一个程序，专门定时下载多个账号的收藏夹视频。

他叫啵啵，所以项目就叫bilibo了

![bobo](./.assets/bobo.JPG)

## 运行步骤
- 复制项目下的`config.yaml.example`为`config.yaml`
- 对应修改配置项
- unix系统：`chmod +x bilibo`，执行`config=config.yaml ./bilibo`运行程序
- 打开配置中的server.host&port，默认：`localhost:8080`

## 配置文件
```yaml
server:                   # web服务配置
  host: 0.0.0.0           # web监听地址
  port: 8080              # web监听端口
  db:                     # 数据库配置，支持sqlite和mysql
    driver: sqlite
    dsn: data.db
    # driver: mysql
    # dsn: user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local

download:                 # 下载配置，目前设置下载路径
  path: /data/downloads
```

## Docker Compose 文件示例
该项目为`Linux/amd64`提供了 Docker 版本镜像。

以下是一个 Docker Compose 的编写示例：
```yaml
services:
  bilibo:
    image: boredtape/bilibo
    volumes:
      - ./data:/app/data
      - ./downloads:/downloads
    restart: always
    container_name: bilibo
    ports:
      - 8080:8080
````

## 同步逻辑
- 收藏夹
  - 视频稿件失效：已下载的视频稿件文件夹会被打上`[已失效]`标识
  - 视频稿件取消收藏：已下载视频保留。未下载视频不下载。
  - 删除收藏夹：已同步的整个收藏夹移动到回收站中
  - 收藏夹改名：如当前收藏夹视频正在下载，等下载完毕后更改收藏夹名称
- 收藏和订阅
  - 视频稿件失效：已下载的视频稿件文件夹会被打上`[已失效]`标识
  - 取消订阅：已同步的整个订阅移动到回收站中
  - 订阅合集中视频稿件被移除合集：保留已下载视频稿件
- 稍后再看
  - 手动删除稍后再看的视频稿件：已下载的视频稿件文件夹移动到回收站中
  - 视频稿件失效：已下载的视频稿件文件夹会被打上`[已失效]`标识

## 预览图
![账号列表](./.assets/1.png)
![设置收藏夹同步](./.assets/2.png)
![已下载视频浏览](./.assets/3.png)
![预览视频](./.assets/4.png)
![同步详情](./.assets/5.png)


## 路线图
- [x] 多账号
- [x] 同步设置
- [x] 视频下载
- [x] 已下载视频浏览
- [x] 预览视频
- [x] 同步详情
- [x] 提供docker分发方式
- [x] 支持`稍后再看`视频下载
- [x] 支持查看当前存在的任务
- [x] 支持`我的收藏和订阅`
- [x] 提供docker除X86以外的镜像
- [ ] 完全不知道接下来可以做什么。。。


## 参考与借鉴(PS:这段都是抄袭bili sync的)

该项目实现过程中主要参考借鉴了如下的项目，感谢他们的贡献：

+ [bili sync](https://github.com/amtoaer/bili-sync) 基于 rust tokio 编写的 bilibili 收藏夹同步下载工具。
+ [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) B 站的第三方接口文档
+ [bilibili-api](https://github.com/Nemo2011/bilibili-api) 使用 Python 调用接口的参考实现
+ [danmu2ass](https://github.com/gwy15/danmu2ass) 本项目弹幕下载功能的缝合来源