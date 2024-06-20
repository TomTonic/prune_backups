package main

import (
	"fmt"
	"time"
)

func main() {

	var date time.Time = time.Date(2024, 5, 31, 23, 54, 21, 0, time.UTC)
	var count int = 30
	var format_print int = 3 // 2=YYYY-MM, 3=YYYY-MM-DD, 4=YYYY-MM-DD_HH, 5=YYYY-MM-DD_HH-mm
	var format_jump int = 3  // 2=month, 3=day, 4=hour
	var per_line int = 6     // line break after ... entries

	for i := 1; i < count+1; i++ {
		fmt.Print("\"")
		switch format_print {
		case 2:
			fmt.Print(date.Format("2006-01"))
		case 3:
			fmt.Print(date.Format("2006-01-02"))
		case 4:
			fmt.Print(date.Format("2006-01-02_15"))
		default:
			fmt.Print(date.Format("2006-01-02_15-04"))
		}
		switch format_jump {
		case 2:
			date = date.AddDate(0, -1, 0)
		case 3:
			date = date.AddDate(0, 0, -1)
		default:
			date = date.Add(-1 * time.Hour)
		}
		if per_line > 1 && i%per_line == 0 {
			fmt.Println("\", ")
		} else {
			fmt.Print("\", ")
		}
	}
	fmt.Println()
}
