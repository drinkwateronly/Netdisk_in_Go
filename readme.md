# 基于Gin框架的网盘后端

在线预览地址：http://119.91.137.20:8081/login
- 测试账号：test1 密码 123456
- 测试账号：test2 密码 123456

首次加载会比较慢


## 前端

前端使用了开源的奇文网盘：

> https://github.com/qiwenshare/qiwen-file-web

前端的本地部署：

1. 拉取前端仓库代码
2. 修改`vue.config.js`配置文件。若使用默认配置，前端将运行在`localhost:8081`，并将后端请求转发到`localhost:8080`
3. 安装npm
4. 进入前端主目录内，执行`npm run serve`，并访问`localhost:8081`检查前端是否部署成功



## 后端

后端的本地部署方式：

1. 拉取本仓库代码
2. 进入MySQL建库建表
   1. 新建数据库：`CREATE DATABASE netdisk;`
   2. 进入数据库：`USE netdisk;`
   3. 执行sql脚本文件建表：`source ./moedls/netdisk.sql;`，其中该sql文件位于项目根目录。
3. 修改配置文件
4. 命令行进入项目目录，执行`go mod tidy`
5. 命令行执行`go run ./main.go`

## swagger
后端部署完毕后，命令行执行`swag init`后，访问 http://localhost:8080/swagger/index.html#/ 即可查看接口文档。


## 已完成功能

- 用户服务
  - 用户注册
  - 用户登录
  - 检查登录状态
- 网盘文件服务
  - 用户文件按路径/类型预览
  - 文件预览
  - 获取用户文件树，用于文件移动/保存分享
  - 文件单个/批量下载
  - 新建文件，文件夹
  - 文件单个/批量删除
  - 文件重命名
  - 文件单个/批量移动
- 文件分享
  - 创建分享连接
  - 分享保存
- 文件回收
  - 回收站文件批量恢复
  - 回收站文件删除
- OnlyOffice
  - office文件预览
  - office文件编辑








## 表设计

在models文件夹下可找到建表的sql文件，各字段可见注释。

### 用户信息表 `user_basic`

```mysql
CREATE TABLE `user_basic` (
  `user_id` char(36) NOT NULL COMMENT '用户标识符',
  `user_type` tinyint unsigned NOT NULL COMMENT '用户类型',
  `username` varchar(64) NOT NULL COMMENT '用户名',
  `password` varchar(64) NOT NULL COMMENT '加盐后的用户密码',
  `phone` varchar(64) NOT NULL COMMENT '用户手机号',
  `total_storage_size` bigint unsigned NOT NULL COMMENT '网盘总容量',
  `storage_size` bigint unsigned NOT NULL COMMENT '网盘已用容量',
  `salt` char(36) NOT NULL COMMENT '盐',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```

### 中心存储池表 `repository_pool`

管理所有真实文件（非文件夹）的存储位置

```mysql
CREATE TABLE `repository_pool` (
  `file_id` char(36) NOT NULL COMMENT '文件标识符',
  `hash` char(32) NOT NULL COMMENT '文件哈希',
  `size` bigint unsigned NOT NULL DEFAULT '0' COMMENT '文件大小',
  `path` varchar(256) NOT NULL COMMENT '文件的本地存储地址/对象存储路径',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`file_id`),
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```

### 用户文件表 `user_repository`

记录用户网盘内文件/文件夹信息，维护用户网盘的文件逻辑，用户文件记录通过`file_id`字段指向中心存储池表记录。

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

- 字段`user_file_id`作为主键索引，
- 字段(`parent_id`,`file_name`,`extend_name`,`is_dir`,`deleted_at`)作为唯一联合索引，用于快速定位某文件，并防止用户文件重复











