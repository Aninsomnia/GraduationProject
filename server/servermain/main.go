package servermain

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func Main() {

	// 设置 log 参数 ，打印当前时间 和 当前行数
	log.SetFlags(log.Ltime | log.Llongfile)

	// ETCD 默认端口号是 2379
	// 使用 ETCD 的 clientv3 包
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
		//超时时间 10 秒
		DialTimeout: 10 * time.Second,
	})

	if err != nil {
		log.Printf("connect to etcd error : %v\n", err)
		return
	}

	log.Printf("connect to etcd successfully ...")
	// defer 最后关闭 连接
	defer client.Close()

	// PUT KEY 为 name , value 为 xiaomotong
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err = client.Put(ctx, "name", "lhq")
	cancel()
	if err != nil {
		log.Printf("PUT key to etcd error : %v\n", err)
		return
	}

	// 获取ETCD 的KEY
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	resp, err := client.Get(ctx, "name")
	cancel()
	if err != nil {
		log.Printf("GET key-value from etcd error : %v\n", err)
		return
	}

	// 遍历读出 KEY 和对应的 value
	for _, ev := range resp.Kvs {
		log.Printf("%s : %s\n", ev.Key, ev.Value)
	}

	cmd := exec.Command("/tmp/etcd-download-test/etcdctl", "member", "list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("combined out:\n%s\n", string(out))
}
