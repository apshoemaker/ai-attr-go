package core

import (
	"testing"
)

func TestContains(t *testing.T) {
	if !SingleLine(5).Contains(5) {
		t.Error("Single(5) should contain 5")
	}
	if SingleLine(5).Contains(6) {
		t.Error("Single(5) should not contain 6")
	}
	if !Range(3, 7).Contains(3) {
		t.Error("Range(3,7) should contain 3")
	}
	if !Range(3, 7).Contains(5) {
		t.Error("Range(3,7) should contain 5")
	}
	if !Range(3, 7).Contains(7) {
		t.Error("Range(3,7) should contain 7")
	}
	if Range(3, 7).Contains(2) {
		t.Error("Range(3,7) should not contain 2")
	}
	if Range(3, 7).Contains(8) {
		t.Error("Range(3,7) should not contain 8")
	}
}

func TestOverlaps(t *testing.T) {
	if !SingleLine(5).Overlaps(SingleLine(5)) {
		t.Error("Single(5) should overlap Single(5)")
	}
	if SingleLine(5).Overlaps(SingleLine(6)) {
		t.Error("Single(5) should not overlap Single(6)")
	}
	if !SingleLine(5).Overlaps(Range(3, 7)) {
		t.Error("Single(5) should overlap Range(3,7)")
	}
	if SingleLine(2).Overlaps(Range(3, 7)) {
		t.Error("Single(2) should not overlap Range(3,7)")
	}
	if !Range(1, 5).Overlaps(Range(3, 7)) {
		t.Error("Range(1,5) should overlap Range(3,7)")
	}
	if Range(1, 2).Overlaps(Range(3, 7)) {
		t.Error("Range(1,2) should not overlap Range(3,7)")
	}
}

func TestRemoveSingleFromSingle(t *testing.T) {
	result := SingleLine(5).Remove(SingleLine(5))
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}

	result = SingleLine(5).Remove(SingleLine(3))
	if len(result) != 1 || result[0] != SingleLine(5) {
		t.Errorf("expected [Single(5)], got %v", result)
	}
}

func TestRemoveRangeFromRange(t *testing.T) {
	// Remove middle
	result := Range(1, 10).Remove(Range(4, 6))
	if len(result) != 2 || result[0] != Range(1, 3) || result[1] != Range(7, 10) {
		t.Errorf("expected [Range(1,3), Range(7,10)], got %v", result)
	}

	// Remove start
	result = Range(1, 10).Remove(Range(1, 3))
	if len(result) != 1 || result[0] != Range(4, 10) {
		t.Errorf("expected [Range(4,10)], got %v", result)
	}

	// Remove end
	result = Range(1, 10).Remove(Range(8, 10))
	if len(result) != 1 || result[0] != Range(1, 7) {
		t.Errorf("expected [Range(1,7)], got %v", result)
	}

	// No overlap
	result = Range(1, 5).Remove(Range(7, 10))
	if len(result) != 1 || result[0] != Range(1, 5) {
		t.Errorf("expected [Range(1,5)], got %v", result)
	}
}

func TestCompressLines(t *testing.T) {
	result := CompressLines(nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}

	result = CompressLines([]uint32{5})
	if len(result) != 1 || result[0] != SingleLine(5) {
		t.Errorf("expected [Single(5)], got %v", result)
	}

	result = CompressLines([]uint32{1, 2, 3})
	if len(result) != 1 || result[0] != Range(1, 3) {
		t.Errorf("expected [Range(1,3)], got %v", result)
	}

	result = CompressLines([]uint32{1, 2, 3, 7, 8, 15})
	if len(result) != 3 || result[0] != Range(1, 3) || result[1] != Range(7, 8) || result[2] != SingleLine(15) {
		t.Errorf("expected [Range(1,3), Range(7,8), Single(15)], got %v", result)
	}
}

func TestExpand(t *testing.T) {
	result := SingleLine(5).Expand()
	if len(result) != 1 || result[0] != 5 {
		t.Errorf("expected [5], got %v", result)
	}

	result = Range(3, 6).Expand()
	expected := []uint32{3, 4, 5, 6}
	if len(result) != len(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("index %d: expected %d, got %d", i, v, result[i])
		}
	}
}

func TestDisplay(t *testing.T) {
	if s := SingleLine(5).String(); s != "5" {
		t.Errorf("expected '5', got '%s'", s)
	}
	if s := Range(3, 7).String(); s != "3-7" {
		t.Errorf("expected '3-7', got '%s'", s)
	}
}

func TestFormatLineRanges(t *testing.T) {
	ranges := []LineRange{Range(19, 222), SingleLine(1), SingleLine(2)}
	result := FormatLineRanges(ranges)
	if result != "1,2,19-222" {
		t.Errorf("expected '1,2,19-222', got '%s'", result)
	}
}

func TestParseLineRanges(t *testing.T) {
	ranges, err := ParseLineRanges("1,2,19-222")
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 3 {
		t.Fatalf("expected 3 ranges, got %d", len(ranges))
	}
	if ranges[0] != SingleLine(1) {
		t.Errorf("expected Single(1), got %v", ranges[0])
	}
	if ranges[1] != SingleLine(2) {
		t.Errorf("expected Single(2), got %v", ranges[1])
	}
	if ranges[2] != Range(19, 222) {
		t.Errorf("expected Range(19,222), got %v", ranges[2])
	}
}
