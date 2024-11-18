# Blueprint

![Blueprint](Blueprint.png)

Documentation | [中文文档](README_zh.md)

## Introduction

Blueprint is a lightweight database migration CLI tool written in Go, inspired by [Laravel Migrations](https://laravel.com/docs/master/migrations).

If you are an independent developer or a startup team, tools like Bytebase may feel a little heavy, but Blueprint will be your powerful assistant😎

## Quick start

### Install
Visit the [release page](https://github.com/YianAndCode/blueprint/releases) and download the corresponding binary file.

Then move the downloaded file to any directory included in the `PATH` environment variable (e.g., `/usr/local/bin/blueprint`), and run the following command:

```bash
blueprint help
```

If the setup is successful, you will see the help message output.

### Init a repo

Navigate to your project migrations directory and run the following command:

```bash
blueprint init
```

And this directory will be initialized as a Blueprint repository for storing `.sql` files.

### Create migration

```bash
blueprint create user
# or specify table name later:
blueprint create

# and you can use update for a update migration:
blueprint update
```

The `create` command wille create a pair of `.sql` files suce as `202411181653_create_user.sql` and `202411181653_create_user_rollback.sql`, and you just need to edit these files for your migration.

And `update` command is similar to `create`, the diffrence between them is just the `.sql` filename.

### Run migration

```bash
blueprint
# or
blueprint run
```

Blueprint will executes all `.sql` files those not executed before, and these files will have same batch number.

### Rollback migration

```bash
# to rollback the last batch of migrations, run the following command:
blueprint rollback
# or
blueprint rollback --batch 1

# or rollback the last step of migrations:
blueprint rollback --step 1
```

You can specify how many step(s) you need to rollback, one step means a `.sql` file.

Or you can specify how many batch(es) you need to rollback, the `batch` is introduced at `Run migration`

Only one of `--step` or `--batch` can be specified at a time, default is `--batch 1`
