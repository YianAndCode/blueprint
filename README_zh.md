# Blueprint

![Blueprint](Blueprint.png)

[Documentation](README.md) | 中文文档

## 介绍

Blueprint 是一款由 Go 编写的轻量级数据库 Migration 命令行工具，灵感来源于 [Laravel Migrations](https://laravel.com/docs/master/migrations)。

如果你是独立开发者/初创团队，Bytebase 之类的工具可能会有点重，但 Blueprint 会是你的得力助手😎

## 快速上手

### 安装

执行以下命令：

```bash
go install github.com/YianAndCode/blueprint@latest
```

### 初始化仓库

切换到项目的 Migration 目录，然后执行：

```bash
blueprint init
```

这个目录就会被初始化为保存 `.sql` 文件的 Blueprint 仓库了。

### 创建 Migration

```bash
blueprint create user
# 或者稍后再指定表名
blueprint create

# 也可以用 update 命令来生成一组更新类型的 migration:
blueprint update
```

`create` 命令会创建一组 `.sql` 文件，形如：`202411181653_create_user.sql` 和 `202411181653_create_user_rollback.sql`，接下来你只需要在这一对文件中编写你的 migration 语句。

`update` 命令和 `create` 是一样的，它们的区别只是 `.sql` 文件的名字。

### 执行 Migration

```bash
blueprint
# 或者
blueprint run
```

Blueprint 会执行全部未执行的 `.sql` 文件，并且这些文件的批次号（`batch number`）是相同的。

### 回滚 Migration

```bash
# 回滚最近一批 migrations，执行：
blueprint rollback
# 或者
blueprint rollback --batch 1

# 回滚最后一步 migration：
blueprint rollback --step 1
```

你可以通过 `--step` 指定要回滚多少步，一个 `.sql` 文件表示“一步”；

你也可以通过 `--batch` 指定要回滚多少批，`批次`（`batch`）的概念见`执行 Migration`。

如果不指定参数，默认是 `--batch 1`；`--step` 和 `--batch` 只能指定一个。