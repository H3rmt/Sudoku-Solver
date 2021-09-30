package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

const Size = 9
const CellSize = 3

const IndentAdd = "| "

var Log = false

type Sudoku struct {
	Raw    [][]uint8
	Rows   [][]Tile
	Solver SudokuSolver
}

type SudokuSolver struct {
	Rows  [][]*Tile
	Cols  [][]*Tile
	Cells [][]*Tile
}

type Tile struct {
	val uint8
}

func log(indent string, log ...interface{}) {
	if Log {
		fmt.Print(append([]interface{}{indent}, log...)...)
	}
}

func logln(indent string, log ...interface{}) {
	if Log {
		fmt.Print(append(append([]interface{}{indent}, log...), "\n")...)
	}
}

func (sudoku *Sudoku) createSolver() {
	sudoku.Rows = make([][]Tile, Size)
	for rowindex, row := range sudoku.Raw {
		var newRow = make([]Tile, Size)
		for columnindex := range row {
			newRow[columnindex] = Tile{sudoku.Raw[rowindex][columnindex]}
		}
		sudoku.Rows[rowindex] = newRow
	}
	
	sudoku.Solver.Rows = make([][]*Tile, Size)
	sudoku.Solver.Cols = make([][]*Tile, Size)
	for rind, row := range sudoku.Rows {
		var newRow = make([]*Tile, Size)
		var newColumn = make([]*Tile, Size)
		for columnindex := range row {
			newRow[columnindex] = &sudoku.Rows[rind][columnindex]
			newColumn[columnindex] = &sudoku.Rows[columnindex][rind]
		}
		sudoku.Solver.Rows[rind] = newRow
		sudoku.Solver.Cols[rind] = newColumn
	}
	
	sudoku.Solver.Cells = [][]*Tile{}
	for cell := 0; cell < Size; cell++ {
		var yOffset = (cell / CellSize) * CellSize
		var xOffset = (cell - yOffset) * CellSize
		
		var newCell []*Tile
		for row := 0; row < CellSize; row++ {
			for col := 0; col < CellSize; col++ {
				newCell = append(newCell, &sudoku.Rows[yOffset+row][xOffset+col])
			}
		}
		sudoku.Solver.Cells = append(sudoku.Solver.Cells, newCell)
	}
}

func (solver *SudokuSolver) getNextEmpty(startRow uint8, startColumn uint8) (row uint8, col uint8) {
	if startRow >= Size || startColumn >= Size {
		return Size, Size
	}
	if (*solver.Rows[startRow][startColumn]).val == 0 {
		return startRow, startColumn
	} else {
		return solver.getNextEmpty(getNextRow(startRow, startColumn), getNextCol(startColumn))
	}
}

func (solver *SudokuSolver) getAvailableNumbers(row []*Tile, col []*Tile, cell []*Tile, indent string) (available []uint8) {
	unavailable := make([]bool, Size+1) // +1 because we have values from 0-Size so Size+1 indexes in array are needed
	
	for _, el := range row {
		unavailable[(*el).val] = true
	}
	for _, el := range col {
		unavailable[(*el).val] = true
	}
	for _, el := range cell {
		unavailable[(*el).val] = true
	}
	
	log(indent, unavailable, "  ")
	for _, el2 := range row {
		log("", *el2)
	}
	log("  ")
	for _, el2 := range col {
		log("", *el2)
	}
	log("  ")
	for _, el2 := range cell {
		log("", *el2)
	}
	logln("")
	
	for n := 1; n <= Size; n++ {
		if unavailable[n] == false {
			available = append(available, uint8(n))
		}
	}
	return
}

func getNextRow(row uint8, col uint8) uint8 {
	return row + (col+1)/Size
}

func getNextCol(col uint8) uint8 {
	return (col + 1) % Size
}

func getCell(row uint8, col uint8) uint8 {
	return (row/CellSize)*CellSize + col/CellSize
}

var currTime = time.Now().UnixMilli()
var totalSolveTrys uint64 = 0

func (solver *SudokuSolver) solve(startRow uint8, startColumn uint8, depth uint16) bool {
	logln("")
	indent := ""
	for i := uint16(0); i < depth; i++ {
		indent += IndentAdd
	}
	totalSolveTrys++
	
	var row, col = solver.getNextEmpty(startRow, startColumn)
	
	if row == Size && col == Size {
		logln(indent, "finished")
		solver.print(indent, int(row), int(col))
		return true
	}
	
	logln(indent, "depth:", depth)
	logln(indent, "current cell: ", row, "-", col)
	
	availableValues := solver.getAvailableNumbers(solver.Rows[row], solver.Cols[col], solver.Cells[getCell(row, col)], indent)
	
	logln(indent, "availableValues:", availableValues)
	if len(availableValues) == 0 {
		logln(indent, "availableValues empty")
		solver.print(indent, int(row), int(col))
		return false
	}
	
	solver.print(indent, int(row), int(col))
	
	for _, value := range availableValues {
		logln(indent, "testing value: ", value)
		
		(*solver.Rows[row][col]).val = value
		
		if solver.solve(row, col, depth+1) {
			return true
		}
	}
	(*solver.Rows[row][col]).val = 0
	
	logln("")
	return false
}

func (solver SudokuSolver) print(indent string, currentRow int, currentColumn int) {
	logln(indent, "Total solves:", totalSolveTrys)
	
	log(indent, "Rows:", "\n", indent)
	for rowindex, row := range solver.Rows {
		for colindex, el2 := range row {
			if rowindex == currentRow && colindex == currentColumn {
				log(" ", "╳")
			} else if (*el2).val != 0 {
				log(" ", (*el2).val)
			} else {
				log(" ", "~")
			}
		}
		log("\n", indent)
	}
	logln("")
}

// https://www.surfpoeten.de/apps/sudoku/generator/
// https://www.youtube.com/watch?v=VPVtlODPdPY
func main() {
	var sudoku Sudoku
	
	data, err := ioutil.ReadFile("resources/sudoku2.json")
	if err != nil {
		panic(err)
	}
	
	err = json.Unmarshal(data, &sudoku.Raw)
	if err != nil {
		panic(err)
	}
	
	sudoku.createSolver()
	solved := sudoku.Solver.solve(0, 0, 0)
	
	var elapsedTime = time.Now().UnixMilli() - currTime
	
	if solved {
		Log = true
		logln("\n", "solved Sudoku in ", totalSolveTrys, " Tries  ", elapsedTime/1000, "s ", elapsedTime%1000, "ms")
		sudoku.Solver.print("", Size, Size)
	} else {
		Log = true
		logln("\n", "couldn't solve Sudoku in ", totalSolveTrys, " Tries  ", elapsedTime/1000, "s ", elapsedTime%1000, "ms")
		sudoku.Solver.print("", Size, Size)
	}
}
