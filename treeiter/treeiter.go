//go:build !solution

package treeiter

type TreeNode[T any] interface {
	comparable
	Left() T
	Right() T
}

func DoInOrder[N TreeNode[N]](root N, do func(N)) {
	var zero N
	if root == zero {
		return
	}

	DoInOrder(root.Left(), do)
	do(root)
	DoInOrder(root.Right(), do)
}
