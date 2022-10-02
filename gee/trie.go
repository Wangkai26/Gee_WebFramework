package gee

import (
	"fmt"
	"strings"
)

type node struct {
	pattern  string		//待匹配路由，完整路由
	part     string		//路由一部分
	children []*node	//子节点切片
	isWild   bool		//是否模糊匹配
}

// 实现 node 的字符串打印
func (n *node) String() string {
	return fmt.Sprintf("node{pattern=%s, part=%s, isWild=%t}", n.pattern, n.part, n.isWild)
}

//实现Trie树的节点插入功能
//递归查找每一层的节点，若没有匹配到当前part的节点，则新建，
//匹配结束时，我们可以使用 n.patern=""判断路由规则是否匹配成功
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	//若节点不存在，则新建节点
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

//实现 Trie 数的节点查询功能
//递归查询每一层的节点，退出规则是，匹配到 *，匹配失败，或匹配到了第 len(parts)层节点
//匹配路由，在当前 height，若匹配成功，返回该节点，失败返回 nil
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		// 若匹配到最后一层 或 模糊匹配后
		// 该节点的 pattern 为空，说明匹配失败
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	// 在该层寻找所有可以匹配成功的节点，递归地进行下一层路由的匹配
	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}

// travel，我的理解是查找所有完整路由，不 ok
func (n *node) travel(list *([]*node)) {
	if n.pattern != "" {
		*list = append(*list, n)
	}
	for _, child := range n.children {
		child.travel(list)
	}
}

//查找第一个匹配成功的节点，用于插入，生成前缀树
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

//所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}
