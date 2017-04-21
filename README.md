替代JWT的Token算法，因为使用了Go的gob包进行序列化，因此只能在Go语言中解析，相比JWT减少了Token字符串的长度和生成时间

用法见 /test/example_test.go
可使在 /test 目录下使用 go test -v 查看测试结果