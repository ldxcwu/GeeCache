![cover](https://images.unsplash.com/photo-1634573826817-27d9e8da08df?ixlib=rb-1.2.1&ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8&auto=format&fit=crop&w=1740&q=80)



# :four_leaf_clover: 1. 为什么要用一致性哈希

## 1.1 避免缓存雪崩

若用固定的哈希映射定位服务器，那么添加活着删除节点时，将导致缓存失效，也就是缓存雪崩



## 1.2 节省存储资源

固定用特定的节点存储特定的缓存，避免了不同节点上存放相同的缓存的副本空间开销，当然也需要缓存备份，不过大部分情况下，普通的缓存不需要全量的副本

# :vertical_traffic_light: 2. 什么是一致性哈希

一致性哈希算法将 key 映射到 2^32 的空间中，将这个数字首尾相连，形成一个环。

- 计算节点/机器(通常使用节点的名称、编号和 IP 地址)的哈希值，放置在环上。
- 计算 key 的哈希值，放置在环上，顺时针寻找到的第一个节点，就是应选取的节点/机器。

![consitentHash](asserts/add_peer.jpg)

环上有 peer2，peer4，peer6 三个节点，`key11`，`key2`，`key27` 均映射到 peer2，`key23` 映射到 peer4。此时，如果新增节点/机器 peer8，假设它新增位置如图所示，那么只有 `key27` 从 peer2 调整到 peer8，其余的映射均没有发生改变。

也就是说，一致性哈希算法，在新增/删除节点时，只需要重新定位该节点附近的一小部分数据，而不需要重新定位所有的节点，这就解决了上述的问题。

### **2.2 数据倾斜问题**

如果服务器的节点过少，容易引起 key 的倾斜。例如上面例子中的 peer2，peer4，peer6 分布在环的上半部分，下半部分是空的。那么映射到环下半部分的 key 都会被分配给 peer2，key 过度向 peer2 倾斜，缓存节点间负载不均。

为了解决这个问题，引入了虚拟节点的概念，一个真实节点对应多个虚拟节点。

假设 1 个真实节点对应 3 个虚拟节点，那么 peer1 对应的虚拟节点是 peer1-1、 peer1-2、 peer1-3（通常以添加编号的方式实现），其余节点也以相同的方式操作。

- 第一步，计算虚拟节点的 Hash 值，放置在环上。
- 第二步，计算 key 的 Hash 值，在环上顺时针寻找到应选取的虚拟节点，例如是 peer2-1，那么就对应真实节点 peer2。

虚拟节点扩充了节点的数量，解决了节点较少的情况下数据容易倾斜的问题。而且代价非常小，只需要增加一个字典(map)维护真实节点与虚拟节点的映射关系即可。

# :v: 3. Golang实现一致性哈希

```go
package geecache

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type Hash func([]byte) uint32

type ConsistentHash struct {
	//hash方法，往哈希环上添加节点或者查找key时所使用的hash方法
	hash Hash
	//每个真实节点所对应的虚拟节点的个数
	virtualNodes int
	//节点数组，包括虚拟节点，对所有节点进行编号并排序，
	nodes []int
	//节点map，通过节点编号，找到真实节点的名字，可以是节点的IP地址或具有意义的App名
	nodesMap map[int]string
	sync.RWMutex
}

func NewConsistent(virtualNodes int, hash Hash) *ConsistentHash {
	if hash == nil {
		hash = crc32.ChecksumIEEE
	}
	return &ConsistentHash{
		hash:         hash,
		virtualNodes: virtualNodes,
		nodesMap:     make(map[int]string),
	}
}

func (c *ConsistentHash) AddNode(node string) {
	//对每个真实节点创建虚拟节点
	for i := 0; i < c.virtualNodes; i++ {
		hashValue := c.hash([]byte(strconv.Itoa(i) + node))
		c.nodes = append(c.nodes, int(hashValue))
		c.nodesMap[int(hashValue)] = node
	}
	sort.Ints(c.nodes)
}

func (c *ConsistentHash) AddNodes(nodes ...string) {
	for _, node := range nodes {
		c.AddNode(node)
	}
}

func (c *ConsistentHash) Get(key string) string {
	if len(c.nodes) == 0 {
		return ""
	}
	hashValue := int(c.hash([]byte(key)))
	//二分搜索
	idx := sort.Search(len(c.nodes), func(i int) bool {
		return c.nodes[i] >= hashValue
	})
	return c.nodesMap[c.nodes[(idx%len(c.nodes))]]
}
//测试程序
package geecache

import (
	"strconv"
	"testing"
)

func TestConsistentHash(t *testing.T) {
	ConsistentRing := NewConsistent(3, func(b []byte) uint32 {
		i, _ := strconv.Atoi(string(b))
		return uint32(i)
	})

	ConsistentRing.AddNodes("2", "4", "6")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, node := range testCases {
		if tmp := ConsistentRing.Get(k); tmp != node {
			t.Errorf("Got wrong node %s, expect %s", tmp, node)
		}
	}

	ConsistentRing.AddNode("8")

	testCases["27"] = "8"

	for k, node := range testCases {
		if tmp := ConsistentRing.Get(k); tmp != node {
			t.Errorf("Got wrong node %s, expect %s", tmp, node)
		}
	}
}
```