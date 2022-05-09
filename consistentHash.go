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
