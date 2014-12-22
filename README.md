zeropush
========
[![GoDoc](http://godoc.org/github.com/sinangedik/zeropush?status.svg)](http://godoc.org/github.com/sinangedik/zeropush)
[![Build Status](https://travis-ci.org/sinangedik/zeropush.svg?branch=master)](https://travis-ci.org/sinangedik/zeropush)

Go client for the ZeroPush API.

I wrote a Go client for [zeropush](https://zeropush.com) since I needed it in a project. I know they are improving their REST API so the client needs more brushing.

I used [Gomega](http://onsi.github.io/gomega/) and [Ginkgo](http://onsi.github.io/ginkgo/) as the [BDD](http://guide.agilealliance.org/guide/bdd.html) framework.

USAGE
========
Here is the GoDoc for the library:

http://godoc.org/github.com/sinangedik/zeropush

Example usage:

First, set your API tokens as environment variables:

```sh
ZEROPUSH_DEV_TOKEN = your_dev_token
ZEROPUSH_PROD_TOKEN = your_prod_token
```

```go
//Initialize the client
zeropushClient := zeropush.NewClient()
//send the notification
_, _ = zeropushClient.Notify("@somebody started following you", "1",  "Tock.tiff", `{"key1" : "value1", "key2", "value2"}`, "", "", "LikeNotification", "your_device_token")
```

TODO
========
Better GoDoc
