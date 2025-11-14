
## Postgres 常用命令整理

### 启动/停止/重启服务
```sh
# 启动Postgres服务
sudo systemctl start postgresql
# 停止服务
sudo systemctl stop postgresql
# 重启服务
sudo systemctl restart postgresql
```

### 登录数据库
```sh
psql -U <用户名> -d <数据库名> -h <主机> -p <端口>
# 例如
psql -U postgres -d testdb -h localhost -p 5432
```

### 数据库操作
```sh
# 创建数据库
CREATE DATABASE dbname;
# 删除数据库
DROP DATABASE dbname;
# 切换数据库
\c dbname
```

### 用户与权限
```sh
# 创建用户
CREATE USER username WITH PASSWORD 'password';
# 授权
GRANT ALL PRIVILEGES ON DATABASE dbname TO username;
# 修改用户密码
ALTER USER username WITH PASSWORD 'newpassword';
```

### 表操作
```sh
# 创建表
CREATE TABLE tablename (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL
);
# 删除表
DROP TABLE tablename;  # 如果表存在依赖，需加CASCADE
DROP TABLE IF EXISTS tablename CASCADE;  # 推荐用法，强制删除表及依赖对象
# 查看表结构
\d tablename

### 修改字段（列）
```sh
# 添加字段
ALTER TABLE tablename ADD COLUMN new_column datatype;
# 删除字段
ALTER TABLE tablename DROP COLUMN column_name;
# 修改字段类型
ALTER TABLE tablename ALTER COLUMN column_name TYPE new_datatype;
# 修改字段名
ALTER TABLE tablename RENAME COLUMN old_name TO new_name;
# 设置/取消字段默认值
ALTER TABLE tablename ALTER COLUMN column_name SET DEFAULT default_value;
ALTER TABLE tablename ALTER COLUMN column_name DROP DEFAULT;
# 设置字段为 NOT NULL / 允许 NULL
ALTER TABLE tablename ALTER COLUMN column_name SET NOT NULL;
ALTER TABLE tablename ALTER COLUMN column_name DROP NOT NULL;
```
```

### 数据操作
```sh
# 插入数据
INSERT INTO tablename (name) VALUES ('value');
# 查询数据
SELECT * FROM tablename;
# 更新数据
UPDATE tablename SET name='newvalue' WHERE id=1;
# 删除数据
DELETE FROM tablename WHERE id=1;
```

### 备份与恢复
```sh
# 备份数据库
pg_dump -U <用户名> -h <主机> -p <端口> <数据库名> > backup.sql
# 恢复数据库
psql -U <用户名> -d <数据库名> -f backup.sql
```

### 查看连接/进程
```sh
# 查看当前连接
SELECT * FROM pg_stat_activity;
# 终止连接
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='dbname' AND pid<>pg_backend_pid();
```

### 其他常用命令
```sh
# 列出所有数据库
\l
# 列出所有表
\dt
# 查看当前用户
SELECT current_user;
# 查看Postgres版本
SELECT version();
```

---
如需更多命令可参考官方文档：https://www.postgresql.org/docs/
