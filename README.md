# SSM Param Store

[![Build Status](https://travis-ci.org/sthulb/ssm-param-store.svg?branch=master)](https://travis-ci.org/sthulb/ssm-param-store)
[![GitHub license](https://img.shields.io/github/license/sthulb/ssm-param-store.svg)](https://github.com/sthulb/ssm-param-store/blob/master/LICENSE)

Wraps the AWS Go SDK's Parameter Store into a higher level package. Adding refreshing of values based on expiries.


## How to Install

Install the package:
```go
go get github.com/sthulb/ssm-param-store
```

## How to use

Simple use:
```go
package main

import "fmt"
import "github.com/sthulb/ssm-param-store"

func main() {
    ps := store.New()
    param, err := ps.Param("name")
    if err != nil {
        panic(err)
    }

    fmt.Print(param.StringValue())
}
```