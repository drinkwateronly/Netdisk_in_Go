# 基于go-gin+gorm的网盘后端系统

基于go-gin+gorm的网盘后端系统，使用了开源的奇文网盘作为前端。前端仓库地址：https://github.com/qiwenshare/qiwen-file-web

本项目在线预览地址：http://119.91.137.20:8081/login

- 测试账号：test1 密码 123456
- 测试账号：test2 密码 123456

首次加载会比较慢



## 功能

基本是按照奇文网盘的前端，实现了其大部分功能，大致包括：

- 用户服务
  - 用户注册
  - 用户登录
  - 检查登录状态
- 网盘文件服务
  - 按文件路径/文件类型预览文件
  - 文件/文件夹新建
  - 文件/文件夹的分片上传
  - 文件秒传
  - 文件单个/批量下载
    - 其中，文件批量下载包括『多个文件下载』或『单个文件夹下载』，此时下载批量文件压缩为zip形式。
  - 文件单个/批量删除
  - 文件单个/批量移动
  - 文件单个/批量复制
  - 图像音视频文件的在线预览，视频支持进度条拖动。
  - 文件重命名
  - 获取用户文件树，用于文件移动/保存分享/文件复制时选择存放的文件夹
- 文件分享
  - 创建分享连接，支持过期时间的设置
  - 分享链接内的文件预览
  - 分享保存
  - 用户已分享的多次文件预览
- 文件回收
  - 回收站文件预览
  - 回收站文件批量恢复
  - 回收站文件单个/批量删除
- OnlyOffice
  - office文件预览
  - office文件编辑



## 本地部署

### 前端

前端的本地部署：

1. 拉取奇文网盘前端仓库代码
2. 修改`vue.config.js`配置文件。若使用默认配置，前端将运行在`localhost:8081`，并将后端请求转发到`localhost:8080`
3. 安装npm
4. 命令行进入前端项目根目录，执行`npm run serve`，并访问`localhost:8081`检查前端是否部署成功



### 后端

后端的本地部署方式：

1. 拉取本仓库代码
2. 进入MySQL建库建表
   1. 新建数据库：`CREATE DATABASE netdisk;`
   2. 进入数据库：`USE netdisk;`
   3. 执行sql脚本文件建表：`SOURCE ./moedls/netdisk.sql;`，其中该sql文件位于本项目根目录。
3. 安装配置OnlyOffice（可选）
4. 修改配置文件`./config.yaml`
5. 命令行进入项目根目录，执行`go mod tidy`
6. 命令行执行`go run ./main.go`



## swagger

后端部署完毕后，命令行执行`swag init`后，访问 http://localhost:8080/swagger/index.html#/ 即可查看各接口文档。

<img src="https://raw.githubusercontent.com/drinkwateronly/Image-Host/main/image/image-20240417204041020.png" alt="image-20240417204041020" style="zoom:67%;" />




## 表设计

![image-20240417203530275](https://raw.githubusercontent.com/drinkwateronly/Image-Host/main/image/image-20240417203530275.png)

### 用户文件表 `user_repository`

其中，较为重要的表为用户文件表，用于维护用户网盘的树型基本逻辑，记录用户网盘内文件/文件夹信息。
- 字段`user_file_id`作为主键索引，
- 字段(`parent_id`,`file_name`,`extend_name`,`is_dir`,`deleted_at`)作为唯一联合索引，用于快速定位某文件，并防止用户文件重复

```mysql
CREATE TABLE `user_repository` (
 `user_id` char(36) NOT NULL COMMENT '文件所有者标识符',
 `user_file_id` char(36) NOT NULL COMMENT '用户文件标识符，唯一索引',
 `file_id` char(36) NOT NULL DEFAULT '' COMMENT '中心存储文件标识符',
 `parent_id` char(36) NOT NULL COMMENT '父文件夹id',
 `file_path` varchar(512) NOT NULL COMMENT '父文件夹绝对路径',
 `file_name` varchar(128) NOT NULL COMMENT '文件名全名',
 `extend_name` varchar(32) NOT NULL DEFAULT '' COMMENT '文件拓展名',
 `file_type` tinyint unsigned NOT NULL COMMENT '文件类型',
 `is_dir` tinyint unsigned NOT NULL COMMENT '是否是文件夹',
 `file_size` bigint unsigned DEFAULT 0 NOT NULL COMMENT '文件大小',
 `modify_time` varchar(64) NOT NULL COMMENT '文件修改时间',
 `upload_time` varchar(64) NOT NULL COMMENT '文件上传时间',
 `created_at` datetime(3) NOT NULL,
 `updated_at` datetime(3) NOT NULL,
 `deleted_at` int unsigned NOT NULL DEFAULT '0' COMMENT 'unix时间戳',
 `delete_batch_id` char(36) DEFAULT NULL COMMENT '文件删除的批次',
 PRIMARY KEY (`user_file_id`),
 UNIQUE KEY `idx_unique_file` (`parent_id`,`file_name`,`is_dir`,`deleted_at`) USING BTREE COMMENT '多列索引，防止文件重复'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```





## todo

- 目前文件存放在磁盘，考虑引入OSS





