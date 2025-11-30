# libxxd
A library to extend functionality provided by the original project (see below)., Some work to be done yet ...

### Typical Usage
```go
	fileName := "./testdata/hello.txt"

	inFile, err := os.Open(fileName)
	if err != nil {
		t.Error(err)
	}

	buf := &bytes.Buffer{}
	writer := bufio.NewWriter(buf)

	xxdCfg := &xxd.Config{DumpType: 0, AutoSkip: false, Bars: true}
	if err := xxd.XxdBasic(inFile, writer, xxdCfg); err != nil {
		t.Error(err)
	}
	writer.Flush()
	//buf should now have the data


```

### Output

```shell

rbalgi@Raghavendras-iMac libxxd % go test -test.v -test.v ./...
=== RUN   TestXXD
0000000: 6865 6c6c 6f2c 2077 6f72 6c64 2120 5468   hello, world! Th
0000010: 6973 2069 7320 7361 7475 7264 6179 2074   is is saturday t
0000020: 6865 2032 3974 6820 6f66 204e 6f76 656d   he 29th of Novem
0000030: 6265 7220 3230 3235 2e20 4927 6d20 7472   ber 2025. I'm tr
0000040: 7969 6e67 2074 6f20 706f 7274 2074 6869   ying to port thi
0000050: 7320 746f 2061 206c 6962 7261 7279 2073   s to a library s
0000060: 7479 6c65 2070 6163 6b61 6765 2057 6865   tyle package Whe
0000070: 7265 7265 7272 7265 6572 3f3f 3f3f 3f3f   rererrreer??????
0000080: 3f3f 3f3f 3f3f 3f0a 6865 6c6c 6f20 776f   ???????.hello wo
0000090: 720a                                      r.

--- PASS: TestXXD (0.00s)
PASS
ok  	github.com/rkbalgi/libxxd	(cached)
?   	github.com/rkbalgi/libxxd/cmd	[no test files]
rbalgi@Raghavendras-iMac libxxd % xxd ./testdata/hello.txt
00000000: 6865 6c6c 6f2c 2077 6f72 6c64 2120 5468  hello, world! Th
00000010: 6973 2069 7320 7361 7475 7264 6179 2074  is is saturday t
00000020: 6865 2032 3974 6820 6f66 204e 6f76 656d  he 29th of Novem
00000030: 6265 7220 3230 3235 2e20 4927 6d20 7472  ber 2025. I'm tr
00000040: 7969 6e67 2074 6f20 706f 7274 2074 6869  ying to port thi
00000050: 7320 746f 2061 206c 6962 7261 7279 2073  s to a library s
00000060: 7479 6c65 2070 6163 6b61 6765 2057 6865  tyle package Whe
00000070: 7265 7265 7272 7265 6572 3f3f 3f3f 3f3f  rererrreer??????
00000080: 3f3f 3f3f 3f3f 3f0a 6865 6c6c 6f20 776f  ???????.hello wo
00000090: 720a                                     r.
rbalgi@Raghavendras-iMac libxxd %


# go-xxd

This repository contains my answer to [How can I improve the performance of
my xxd
port?](http://www.reddit.com/r/golang/comments/2s1zn1/how_can_i_improve_the_performance_of_my_xxd_port/)
on reddit.

The result is a Go version of xxd that outperforms the native versions on OSX
10.10.1 / Ubuntu 14.04 (inside VirtualBox), see benchmarks below. However, that
is not impressive, given that none of the usual xxd flags are supported.

What is interesting however, are the steps to get there:

* Make the code testable and compare against output of native xxd using test/quick: https://github.com/felixge/go-xxd/commit/90262b3dcdc518ca3eaec7171aa14d74d95f34b8
* Fix bugs: https://github.com/felixge/go-xxd/commit/e9ebeb0abdf78f6e7729fdbfc68842b3a86ee0a3, https://github.com/felixge/go-xxd/commit/120804574f12033999f23e6cf6a3b75961f14da1, https://github.com/felixge/go-xxd/commit/dab678ecf5dcb3eff345db8ac68ae6d7438f9d0e
* Buffer output to stdout: https://github.com/felixge/go-xxd/commit/69b5fe0cc7da80d374413d72892507d5e5ecaabc
* Implement a benchmark: https://github.com/felixge/go-xxd/commit/0bce954073ce92b72ed3fbcf36603c6e23852feb
* Use the benchmark + [go pprof](http://blog.golang.org/profiling-go-programs) to get ideas for optimization
* Remove unneeded printf calls: https://github.com/felixge/go-xxd/commit/473330acd320e5318e896d6408fb3d64a5b8e10b
* Optimize hex encoding and avoid allocations: https://github.com/felixge/go-xxd/commit/dce3bca200ae499e1cf57994c7592a42f66694d5, https://github.com/felixge/go-xxd/commit/a48d892de3fb625bdb3e8e367337578c766a42f5, https://github.com/felixge/go-xxd/commit/d653d2f4eeb5fa41e907f260bd95c205bd8e1ff7, https://github.com/felixge/go-xxd/commit/0d3ae7f0be863a138fab3fd2dd89208073b61c7f
* Avoid type casts: https://github.com/felixge/go-xxd/commit/859ebc4489ce81edc0681881700b0f22754943f0

You can also follow along by looking at the commit history: https://github.com/felixge/go-xxd/commits/master

## OSX 10.10.1:

### xxd native:

```
$ time xxd image.jpg > /dev/null

real	0m0.205s
user	0m0.202s
sys	0m0.003s
```

### xxd.go (original version from reddit):

```
$ go build xxd.go && time ./xxd image.jpg > /dev/null

real	0m5.914s
user	0m3.598s
sys	0m2.318s
```

### xxd.go (optimized):

```
$ go build xxd.go && time ./xxd image.jpg > /dev/null

real	0m0.138s
user	0m0.133s
sys	0m0.004s
```

## Ubuntu 14.04 (inside VirtualBox):

### xxd native:

```
$ time xxd image.jpg > /dev/null

real	0m0.273s
user	0m0.017s
sys	0m0.231s
```

### xxd.go (original version from reddit):

```
$ go build xxd.go && time ./xxd image.jpg > /dev/null

real	0m5.856s
user	0m3.517s
sys	0m1.897s
```

### xxd.go (optimized):

```
$ go build xxd.go && time ./xxd image.jpg > /dev/null

real	0m0.233s
user	0m0.021s
sys	0m0.207s
```
