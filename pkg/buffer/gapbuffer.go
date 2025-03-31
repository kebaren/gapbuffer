package buffer

import (
	"errors"
	"unicode/utf8"
)

const (
	// Use larger chunks to improve performance
	DEFAULT_CHUNK_SIZE = 4096        // 4KB chunks
	DEFAULT_GAP_SIZE   = 1024 * 1024 // 1MB initial gap
)

// Chunk represents a chunk of text in the gap buffer
type Chunk struct {
	Text string
	Pos  int
}

// GapBuffer represents a gap buffer implemented using a red-black tree
type GapBuffer struct {
	tree      *RBTree
	gapStart  int
	gapEnd    int
	length    int
	chunkSize int
}

// New creates a new gap buffer
func New() *GapBuffer {
	tree := NewRBTree()
	return &GapBuffer{
		tree:      tree,
		gapStart:  0,
		gapEnd:    DEFAULT_GAP_SIZE,
		length:    0,
		chunkSize: DEFAULT_CHUNK_SIZE,
	}
}

// NewWithChunkSize creates a new gap buffer with a specified chunk size
func NewWithChunkSize(chunkSize int) *GapBuffer {
	tree := NewRBTree()
	if chunkSize <= 0 {
		chunkSize = DEFAULT_CHUNK_SIZE
	}
	return &GapBuffer{
		tree:      tree,
		gapStart:  0,
		gapEnd:    DEFAULT_GAP_SIZE,
		length:    0,
		chunkSize: chunkSize,
	}
}

// InsertAt inserts text at the specified position
func (gb *GapBuffer) InsertAt(pos int, text string) error {
	if pos < 0 || pos > gb.length {
		return errors.New("position out of range")
	}

	// Move gap to insertion point if needed
	if pos != gb.gapStart {
		gb.moveGap(pos)
	}

	// If gap is too small, expand it
	if len(text) > gb.gapEnd-gb.gapStart {
		gb.expandGap(len(text))
	}

	// Insert text into gap in chunks, ensuring we don't break Unicode characters
	for i := 0; i < len(text); {
		// Determine end position for this chunk, ensuring we don't break a UTF-8 sequence
		end := i + gb.chunkSize
		if end > len(text) {
			end = len(text)
		} else {
			// Ensure we don't break a UTF-8 sequence
			for !utf8.RuneStart(text[end-1]) && end > i {
				end--
			}
		}

		chunk := &Chunk{
			Text: text[i:end],
			Pos:  gb.gapStart,
		}
		gb.tree.Insert(gb.gapStart, chunk)
		gb.gapStart += end - i

		i = end
	}

	// Update length
	gb.length += len(text)

	return nil
}

// InsertRuneAt 在指定的Unicode字符位置插入文本
func (gb *GapBuffer) InsertRuneAt(runePos int, text string) error {
	// 获取完整文本以计算字节位置
	fullText := gb.GetText()

	// 计算字节位置
	bytePos := RuneIndex(fullText, runePos)
	if bytePos < 0 {
		return errors.New("rune position out of range")
	}

	// 调用字节位置的插入方法
	return gb.InsertAt(bytePos, text)
}

// DeleteAt deletes text at the specified position
func (gb *GapBuffer) DeleteAt(pos int, count int) error {
	if pos < 0 || pos+count > gb.length {
		return errors.New("position or count out of range")
	}

	// Move gap to deletion point if needed
	if pos != gb.gapStart {
		gb.moveGap(pos)
	}

	// Find all chunks that need to be deleted
	var keysToDelete []int
	gb.tree.InOrderTraversal(func(key int, value interface{}) {
		if key >= gb.gapEnd && key < gb.gapEnd+count {
			keysToDelete = append(keysToDelete, key)
		}
	})

	// Delete the chunks
	for _, key := range keysToDelete {
		gb.tree.Delete(key)
	}

	// Update gap and length
	gb.gapEnd += count
	gb.length -= count
	return nil
}

// DeleteRuneAt 删除从指定Unicode字符位置开始的指定数量的Unicode字符
func (gb *GapBuffer) DeleteRuneAt(runePos int, runeCount int) error {
	if runeCount <= 0 {
		return errors.New("rune count must be positive")
	}

	// 获取完整文本以计算字节位置
	fullText := gb.GetText()

	// 确保runePos有效
	if runePos < 0 || runePos >= RuneCount(fullText) {
		return errors.New("rune position out of range")
	}

	// 计算开始和结束的字节位置
	byteStart, byteEnd := RuneIndexRange(fullText, runePos, runePos+runeCount)
	if byteStart < 0 || byteEnd < 0 {
		return errors.New("invalid rune range")
	}

	// 调用字节位置的删除方法
	return gb.DeleteAt(byteStart, byteEnd-byteStart)
}

// GetText returns the text in the buffer
func (gb *GapBuffer) GetText() string {
	// Approximate the buffer size to avoid frequent reallocations
	resultCapacity := gb.length + 100
	if resultCapacity > 100*1024*1024 { // Cap at 100MB to avoid excessive allocation
		resultCapacity = 100 * 1024 * 1024
	}

	// Use a byte buffer for better performance with large strings
	result := make([]byte, 0, resultCapacity)

	gb.tree.InOrderTraversal(func(key int, value interface{}) {
		chunk := value.(*Chunk)
		if key < gb.gapStart || key >= gb.gapEnd {
			result = append(result, chunk.Text...)
		}
	})

	// 确保返回的是有效的UTF-8字符串
	return EnsureValidUTF8(string(result))
}

// GetTextRange returns the text in the specified range
func (gb *GapBuffer) GetTextRange(start int, end int) (string, error) {
	if start < 0 || end > gb.length || start > end {
		return "", errors.New("invalid range")
	}

	length := end - start
	result := make([]byte, 0, length)

	// Convert logical positions to physical positions
	physicalStart := start
	physicalEnd := end
	if physicalStart >= gb.gapStart {
		physicalStart += (gb.gapEnd - gb.gapStart)
	}
	if physicalEnd >= gb.gapStart {
		physicalEnd += (gb.gapEnd - gb.gapStart)
	}

	// Find all chunks in the range and add their text to the result
	gb.tree.InOrderTraversal(func(key int, value interface{}) {
		chunk := value.(*Chunk)
		adjustedKey := key
		chunkText := chunk.Text
		chunkLen := len(chunkText)

		// Skip the gap
		if key >= gb.gapStart && key < gb.gapEnd {
			return
		}

		// Adjust key for logical position calculation
		if key >= gb.gapEnd {
			adjustedKey = key - (gb.gapEnd - gb.gapStart)
		}

		// Check if this chunk is in our range
		if adjustedKey+chunkLen <= start || adjustedKey >= end {
			return
		}

		// Calculate the intersection of the chunk and our range
		chunkStart := 0
		if adjustedKey < start {
			chunkStart = start - adjustedKey
		}

		chunkEnd := chunkLen
		if adjustedKey+chunkLen > end {
			chunkEnd = end - adjustedKey
		}

		// Add the relevant portion of the chunk text to our result
		if chunkStart < chunkEnd {
			// Ensure we don't break UTF-8 sequences at the start
			for chunkStart < chunkEnd && !utf8.RuneStart(chunkText[chunkStart]) {
				chunkStart++
			}

			// Ensure we don't break UTF-8 sequences at the end
			tmpEnd := chunkEnd
			for tmpEnd > chunkStart {
				runeLen := 1
				for i := 1; i < utf8.UTFMax && tmpEnd-i >= chunkStart; i++ {
					if utf8.RuneStart(chunkText[tmpEnd-i]) {
						runeLen = i
						break
					}
				}

				if tmpEnd-runeLen >= chunkStart {
					valid := true
					for i := 0; i < runeLen; i++ {
						if !utf8.ValidRune(rune(chunkText[tmpEnd-runeLen+i])) {
							valid = false
							break
						}
					}

					if valid {
						break
					}
				}

				tmpEnd--
			}

			chunkEnd = tmpEnd

			// Add the valid portion to result
			if chunkStart < chunkEnd {
				result = append(result, chunkText[chunkStart:chunkEnd]...)
			}
		}
	})

	// 确保返回的是有效的UTF-8字符串
	return EnsureValidUTF8(string(result)), nil
}

// GetRuneTextRange 获取指定Unicode字符范围的文本
func (gb *GapBuffer) GetRuneTextRange(runeStart int, runeEnd int) (string, error) {
	if runeStart < 0 || runeStart > runeEnd {
		return "", errors.New("invalid rune range")
	}

	// 获取完整文本以计算字节位置
	fullText := gb.GetText()

	// 确保runeEnd不超出范围
	runeCount := RuneCount(fullText)
	if runeEnd > runeCount {
		runeEnd = runeCount
	}

	// 计算开始和结束的字节位置
	byteStart, byteEnd := RuneIndexRange(fullText, runeStart, runeEnd)
	if byteStart < 0 || byteEnd < 0 {
		return "", errors.New("invalid rune range")
	}

	// 调用字节位置的范围获取方法
	return gb.GetTextRange(byteStart, byteEnd)
}

// Replace replaces the text in the specified range
func (gb *GapBuffer) Replace(start int, end int, text string) error {
	if start < 0 || end > gb.length || start > end {
		return errors.New("invalid range")
	}

	// Delete the range
	if err := gb.DeleteAt(start, end-start); err != nil {
		return err
	}

	// Insert the new text
	return gb.InsertAt(start, text)
}

// ReplaceRune 替换指定Unicode字符范围的文本
func (gb *GapBuffer) ReplaceRune(runeStart int, runeEnd int, text string) error {
	if runeStart < 0 || runeStart > runeEnd {
		return errors.New("invalid rune range")
	}

	// 获取完整文本以计算字节位置
	fullText := gb.GetText()

	// 确保runeEnd不超出范围
	runeCount := RuneCount(fullText)
	if runeEnd > runeCount {
		runeEnd = runeCount
	}

	// 计算开始和结束的字节位置
	byteStart, byteEnd := RuneIndexRange(fullText, runeStart, runeEnd)
	if byteStart < 0 || byteEnd < 0 {
		return errors.New("invalid rune range")
	}

	// 调用字节位置的替换方法
	return gb.Replace(byteStart, byteEnd, text)
}

// moveGap moves the gap to the specified position
func (gb *GapBuffer) moveGap(pos int) {
	if pos == gb.gapStart {
		return
	}

	// Collect nodes that need to be moved
	var nodesToMove []*Node

	if pos < gb.gapStart {
		// Move gap left
		gb.tree.InOrderTraversal(func(key int, value interface{}) {
			if key >= pos && key < gb.gapStart {
				nodesToMove = append(nodesToMove, gb.tree.Search(key))
			}
		})

		// Process nodes from right to left to maintain order
		for i := len(nodesToMove) - 1; i >= 0; i-- {
			node := nodesToMove[i]
			chunk := node.Value.(*Chunk)
			gb.tree.Delete(node.Key)
			newKey := gb.gapEnd - len(nodesToMove) + i
			gb.tree.Insert(newKey, chunk)
		}

		// Update gap boundaries
		gapSize := gb.gapEnd - gb.gapStart
		gb.gapEnd = pos + gapSize
		gb.gapStart = pos

	} else {
		// Move gap right
		gb.tree.InOrderTraversal(func(key int, value interface{}) {
			if key >= gb.gapEnd && key < gb.gapEnd+(pos-gb.gapStart) {
				nodesToMove = append(nodesToMove, gb.tree.Search(key))
			}
		})

		// Process nodes from left to right to maintain order
		for i, node := range nodesToMove {
			chunk := node.Value.(*Chunk)
			gb.tree.Delete(node.Key)
			newKey := gb.gapStart + i
			gb.tree.Insert(newKey, chunk)
		}

		// Update gap boundaries
		gapSize := gb.gapEnd - gb.gapStart
		gb.gapStart = pos
		gb.gapEnd = pos + gapSize
	}
}

// expandGap expands the gap to accommodate more text
func (gb *GapBuffer) expandGap(minSize int) {
	currentGapSize := gb.gapEnd - gb.gapStart
	if currentGapSize >= minSize {
		return
	}

	// Calculate new gap size (double the current size or requested size, whichever is larger)
	newGapSize := currentGapSize * 2
	if newGapSize < minSize {
		newGapSize = minSize
	}

	// We need to expand by this much
	expandBy := newGapSize - currentGapSize

	// Shift all nodes after the gap
	var nodesToMove []*Node
	gb.tree.InOrderTraversal(func(key int, value interface{}) {
		if key >= gb.gapEnd {
			nodesToMove = append(nodesToMove, gb.tree.Search(key))
		}
	})

	// Delete and reinsert nodes with new positions
	for _, node := range nodesToMove {
		key := node.Key
		value := node.Value
		gb.tree.Delete(key)
		gb.tree.Insert(key+expandBy, value)
	}

	gb.gapEnd += expandBy
}

// Length returns the length of the text in the buffer
func (gb *GapBuffer) Length() int {
	return gb.length
}

// RuneLength 返回缓冲区中Unicode字符的数量
func (gb *GapBuffer) RuneLength() int {
	return RuneCount(gb.GetText())
}

// GapLength returns the current gap length
func (gb *GapBuffer) GapLength() int {
	return gb.gapEnd - gb.gapStart
}

// GapStart returns the current gap start position
func (gb *GapBuffer) GapStart() int {
	return gb.gapStart
}

// GapEnd returns the current gap end position
func (gb *GapBuffer) GapEnd() int {
	return gb.gapEnd
}
