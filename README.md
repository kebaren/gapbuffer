# Gap Buffer with Red-Black Tree

这个项目使用Go语言实现了一个基于红黑树的gap buffer数据结构。Gap buffer是一种常用于文本编辑器的数据结构，用于高效地插入和删除文本。

## 概览

Gap buffer通过在当前光标位置维护一个"间隙"，实现高效的文本插入和删除操作。在这个实现中，我们使用红黑树来存储文本数据，这提供了：

- 插入、删除和搜索操作的O(log n)时间复杂度
- 对于非顺序编辑的高效gap移动
- 对文本片段的快速访问
- 完整的Unicode支持

## 功能特性

- 在任意位置插入文本
- 从任意位置删除文本
- 替换指定范围内的文本
- 检索文本或文本范围
- 自动调整gap大小以高效管理内存
- 完全支持Unicode和多字节字符
- 基于字符(rune)的操作接口

## 实现细节

该实现由两个主要组件组成：

1. **红黑树 (pkg/rbtree/rbtree.go)**: 一个自平衡二叉搜索树，保持O(log n)的高度。
2. **Gap Buffer (pkg/gapbuffer/gapbuffer.go)**: 使用红黑树实现gap buffer功能。
3. **Unicode支持 (pkg/gapbuffer/unicode.go)**: 处理多字节字符的辅助函数。

### 优化特性

- 使用分块存储而非单字符存储，减少树节点数量，提高性能
- 具有自动扩展功能的可变大小的gap
- 对Unicode和多字节字符的特殊处理，确保不会在UTF-8序列中间断开
- 针对大型文件的内存优化

## 使用方法

### 基于字节的操作

```go
// 创建一个新的gap buffer
buffer := gapbuffer.New()

// 插入文本
buffer.InsertAt(0, "Hello, 世界!")

// 获取整个文本
text := buffer.GetText()

// 删除文本
buffer.DeleteAt(7, 5) // 从位置7开始删除5个字节

// 替换文本
buffer.Replace(0, 5, "你好")

// 获取文本范围
textRange, err := buffer.GetTextRange(0, 10)
```

### 基于Unicode字符的操作

```go
// 在字符位置插入文本
buffer.InsertRuneAt(2, "【插入】")

// 删除指定数量的字符
buffer.DeleteRuneAt(5, 3)  // 从第5个字符开始删除3个字符

// 获取字符范围的文本
runeRange, _ := buffer.GetRuneTextRange(0, 5)

// 替换字符范围的文本
buffer.ReplaceRune(10, 15, "「替换文本」")

// 获取文本中Unicode字符的数量
charCount := buffer.RuneLength()
```

## 运行示例

要运行演示gap buffer功能的示例程序：

```bash
go run cmd/gapbuffer/main.go
```

要测试性能（包括对100MB文件的操作）：

```bash
go run cmd/perftest/main.go
```

要测试Unicode支持：

```bash
cd cmd/unicodetest && go run .
```

## 性能

使用红黑树作为底层数据结构提供：

- 插入和删除操作：O(log n)
- Gap移动：O(k log n)，其中k是移动的距离
- 文本检索：最坏情况下O(n log n)

在性能测试中，即使对100MB大小的文件操作，也表现出良好的性能：
- 100MB文本填充：617ms
- 100个随机插入操作：538ms
- 100个随机删除操作：49ms
- 100个随机替换操作：73ms
- 100个随机读取操作：14ms
- 完整缓冲区读取：73ms

## Unicode支持

该实现特别注意处理Unicode文本：
- 确保正确处理多字节字符（如中文、emoji等）
- 在分块存储时避免在UTF-8序列中间断开
- 保持字符完整性的文本范围操作
- 提供基于Unicode字符（而非字节）的操作接口
- 自动修复无效的UTF-8序列

支持的Unicode内容包括：
- 中文、日语、俄语、希伯来语、阿拉伯语等多种语言文字
- Emoji和复杂的Unicode组合（如国旗、肤色修饰符等）
- 零宽连接符(ZWJ)字符序列
- 变体选择符(VS)字符序列
- 混合多语言文本

## 许可证

本项目是开源的，使用MIT许可证发布。 