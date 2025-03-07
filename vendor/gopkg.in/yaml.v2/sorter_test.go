package yaml

import (
    "reflect"
    "sort"
    "testing"
)

// TestKeyListNumeric tests sorting of numeric values including ints, floats, and booleans.
func TestKeyListNumeric(t *testing.T) {
    // Create a keyList with various numeric types: bool, int, float.
    ks := keyList{
        reflect.ValueOf(10),
        reflect.ValueOf(2),
        reflect.ValueOf(2.5),
        reflect.ValueOf(true),
        reflect.ValueOf(false),
    }

    // Sorting using keyList.Less which internally compares using keyFloat and numLess.
    sort.Sort(ks)

    // Expected order:
    // keyFloat(false)=0, keyFloat(true)=1, keyFloat(2)=2, keyFloat(2.5)=2.5, keyFloat(10)=10.
    expected := []float64{0, 1, 2, 2.5, 10}
    for i, v := range ks {
        f, ok := keyFloat(v)
        if !ok {
            t.Errorf("Expected numeric value, got non-numeric at index %d", i)
        }
        if f != expected[i] {
            t.Errorf("At index %d, expected %v, got %v", i, expected[i], f)
        }
    }
}

// TestKeyListString tests sorting of strings with natural ordering.
func TestKeyListString(t *testing.T) {
    // Create a keyList with strings that include both uppercase letters and digits.
    keys := keyList{
        reflect.ValueOf("a10"),
        reflect.ValueOf("a2"),
        reflect.ValueOf("a00"),
        reflect.ValueOf("a0"),
        reflect.ValueOf("A"),
        reflect.ValueOf("a"),
    }

    sort.Sort(keys)

    // The expected order determined by the Less function:
    // "A" < "a" < "a0" < "a00" < "a2" < "a10"
    expected := []string{"A", "a", "a0", "a00", "a2", "a10"}
    for i, v := range keys {
        if v.String() != expected[i] {
            t.Errorf("At index %d, expected %s, got %s", i, expected[i], v.String())
        }
    }
}

// TestKeyListPointer tests sorting of pointers by unwrapping them.
func TestKeyListPointer(t *testing.T) {
    a := 42
    b := 7
    c := 100

    // Create a keyList with pointers to ints.
    ks := keyList{
        reflect.ValueOf(&a),
        reflect.ValueOf(&b),
        reflect.ValueOf(&c),
    }

    sort.Sort(ks)

    // After unwrapping, expected order by int value: 7, 42, 100.
    expected := []int{7, 42, 100}
    for i, v := range ks {
        // Unwrap pointer to get int.
        if v.Kind() == reflect.Ptr && !v.IsNil() {
            v = v.Elem()
        }
        if int(v.Int()) != expected[i] {
            t.Errorf("At index %d, expected %d, got %d", i, expected[i], int(v.Int()))
        }
    }
}

// TestKeyListInterface tests sorting of numeric values wrapped in interfaces.
func TestKeyListInterface(t *testing.T) {
    // Wrap numeric values in interfaces.
    var a interface{} = 3
    var b interface{} = 1
    var c interface{} = 2

    ks := keyList{
        reflect.ValueOf(a),
        reflect.ValueOf(b),
        reflect.ValueOf(c),
    }

    sort.Sort(ks)

    expected := []int{1, 2, 3}
    for i, v := range ks {
        // Unwrap interface if necessary.
        for v.Kind() == reflect.Interface && !v.IsNil() {
            v = v.Elem()
        }
        if int(v.Int()) != expected[i] {
            t.Errorf("At index %d, expected %d, got %d", i, expected[i], int(v.Int()))
        }
    }
}

// TestKeyListNonNumericNonString tests sorting for values that are neither numeric nor string.
func TestKeyListNonNumericNonString(t *testing.T) {
    // Define two dummy structs.
    type dummy struct {
        val int
    }
    a := dummy{10}
    b := dummy{5}

    ks := keyList{
        reflect.ValueOf(a),
        reflect.ValueOf(b),
    }

    // Sorting will rely on the reflect.Kind comparison.
    sort.Sort(ks)

    // Since both are structs (same kind), Less returns false (they are considered equal),
    // so the slice remains unchanged. We simply verify the length.
    if len(ks) != 2 {
        t.Errorf("Expected length 2, got %d", len(ks))
    }
}
// TestKeyListEmpty tests sorting an empty keyList.
func TestKeyListEmpty(t *testing.T) {
    var ks keyList
    sort.Sort(ks)
    if len(ks) != 0 {
        t.Errorf("Expected empty keyList to remain empty, got length %d", len(ks))
    }
}

// TestKeyListSingle tests sorting a keyList with a single element.
func TestKeyListSingle(t *testing.T) {
    ks := keyList{reflect.ValueOf(100)}
    sort.Sort(ks)
    if len(ks) != 1 {
        t.Errorf("Expected keyList of length 1, got %d", len(ks))
    }
    f, ok := keyFloat(ks[0])
    if !ok || f != 100 {
        t.Errorf("Expected numeric value 100, got %v", ks[0])
    }
}

// TestKeyListNilPointer tests sorting when some pointers are nil.
func TestKeyListNilPointer(t *testing.T) {
    var nonNil int = 5
    var nilPtr *int = nil
    ks := keyList{
        reflect.ValueOf(nilPtr),
        reflect.ValueOf(&nonNil),
    }
    sort.Sort(ks)
    // For &nonNil, the sorter unwraps it to int (kind reflect.Int)
    // For nilPtr, since it is nil the sorter does not unwrap it (kind reflect.Ptr).
    // Since reflect.Int (a lower constant value) is less than reflect.Ptr,
    // the non-nil pointer should come first.
    a := ks[0]
    if a.Kind() == reflect.Ptr && !a.IsNil() {
        a = a.Elem()
    }
    if int(a.Int()) != 5 {
        t.Errorf("Expected non-nil pointer value 5 at first index, got %v", a)
    }
}

// TestKeyListNestedPointer tests sorting when values are nested pointers (pointer to pointer) 
// and wrapped in interfaces.  This ensures that multiple levels of indirection are correctly unwrapped.
func TestKeyListNestedPointer(t *testing.T) {
    // Create nested pointers: **int
    var val1 int = 3
    var val2 int = 1
    var val3 int = 2
    p1 := &val1
    p2 := &val2
    p3 := &val3
    // Create pointers to pointers.
    pp1 := &p1
    pp2 := &p2
    pp3 := &p3
    // Create a keyList containing these nested pointers.
    ks := keyList{
        reflect.ValueOf(pp1),
        reflect.ValueOf(pp2),
        reflect.ValueOf(pp3),
    }
    sort.Sort(ks)
    // After unwrapping nested pointers, the expected order is: 1, 2, 3.
    expected := []int{1, 2, 3}
    for i, v := range ks {
        // Unwrap all levels if needed.
        for (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && !v.IsNil() {
            v = v.Elem()
        }
        if int(v.Int()) != expected[i] {
            t.Errorf("At index %d, expected %d, got %d", i, expected[i], int(v.Int()))
        }
    }
}
// TestKeyListMixedNumericTypes tests sorting of numeric values of various types including negative values,
func TestKeyListMixedNumericTypes(t *testing.T) {
    ks := keyList{
        reflect.ValueOf(uint(2)),      // reflect.Uint (numeric value 2)
        reflect.ValueOf(int(2)),       // reflect.Int (numeric value 2)
        reflect.ValueOf(float64(2)),   // reflect.Float64 (numeric value 2)
        reflect.ValueOf(false),        // reflect.Bool (numeric value 0)
        reflect.ValueOf(true),         // reflect.Bool (numeric value 1)
        reflect.ValueOf(int(-10)),     // reflect.Int (numeric value -10)
        reflect.ValueOf(10),           // reflect.Int (numeric value 10)
    }
    sort.Sort(ks)

    // Expected order:
    // 1. int(-10)   => -10
    // 2. false      => 0
    // 3. true       => 1
    // 4. Among the three 2's (same numeric value), order by their reflect.Kind:
    //    - int(2)     => reflect.Int (lower kind value)
    //    - uint(2)    => reflect.Uint
    //    - float64(2) => reflect.Float64
    // 5. int(10)    => 10

    expectedKinds := []reflect.Kind{
        reflect.Int,     // for int(-10)
        reflect.Bool,    // for false
        reflect.Bool,    // for true
        reflect.Int,     // for int(2)
        reflect.Uint,    // for uint(2)
        reflect.Float64, // for float64(2)
        reflect.Int,     // for int(10)
    }

    expectedFloats := []float64{-10, 0, 1, 2, 2, 2, 10}

    for i, v := range ks {
        // Unwrap if v is a pointer or interface.
        for (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && !v.IsNil() {
            v = v.Elem()
        }
        f, ok := keyFloat(v)
        if !ok {
            t.Errorf("Index %d: expected numeric value, got non-numeric", i)
        }
        if f != expectedFloats[i] {
            t.Errorf("Index %d: expected numeric %v, got %v", i, expectedFloats[i], f)
        }
        if v.Kind() != expectedKinds[i] {
            t.Errorf("Index %d: expected kind %v, got %v", i, expectedKinds[i], v.Kind())
        }
    }
}

// TestKeyListNilInterface tests sorting when the keyList contains a nil interface value.
func TestKeyListNilInterface(t *testing.T) {
    var a interface{} = nil
    ks := keyList{
        reflect.ValueOf(a),
        reflect.ValueOf(5),
    }
    sort.Sort(ks)

    if len(ks) != 2 {
        t.Errorf("Expected 2 elements, got %d", len(ks))
    }

    // The first element should be the nil interface (which results in an Invalid reflect.Kind)
    if ks[0].Kind() != reflect.Invalid {
        t.Errorf("Expected first element to be Invalid, got %v", ks[0].Kind())
    }

    // For the second element, unwrap it if needed and verify its value.
    v := ks[1]
    for (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && !v.IsNil() {
        v = v.Elem()
    }
    n, ok := keyFloat(v)
    if !ok || n != 5 {
        t.Errorf("Expected numeric value 5, got %v", n)
    }
}
// TestNumLessPanic tests that numLess panics for non-numeric types.
func TestNumLessPanic(t *testing.T) {
    defer func() {
        if r := recover(); r == nil {
            t.Errorf("Expected panic for non-numeric values in numLess, but no panic occurred")
        }
    }()
    // Create two reflect.Values of a non-numeric type: using complex numbers which are not handled.
    a := reflect.ValueOf(complex(1, 1))
    b := reflect.ValueOf(complex(2, 2))
    _ = numLess(a, b)
}

// TestKeyListMixedTypes tests sorting of a keyList containing mixed non-numeric types
// where one is a string and the other is a struct. Since keyFloat fails on both,
// sorting falls back to comparing their reflect.Kind values.
func TestKeyListMixedTypes(t *testing.T) {
    type dummy struct{ val int }
    ks := keyList{
        reflect.ValueOf("hello"),
        reflect.ValueOf(dummy{val: 10}),
    }
    sort.Sort(ks)

    // According to reflect.Kind ordering, reflect.String is less than reflect.Struct,
    // so the string "hello" should appear first.
    if ks[0].Kind() != reflect.String {
        t.Errorf("Expected first element to be a string, got %v", ks[0].Kind())
    }
    if ks[0].String() != "hello" {
        t.Errorf("Expected first element to be 'hello', got %s", ks[0].String())
    }
    }
// TestKeyListComplexString tests natural sorting for complex strings with embedded numbers.
func TestKeyListComplexString(t *testing.T) {
    keys := keyList{
        reflect.ValueOf("a01b"),
        reflect.ValueOf("a001b"),
        reflect.ValueOf("A1b"),
        reflect.ValueOf("a1B"),
        reflect.ValueOf("a1a"),
        reflect.ValueOf("a1b"),
    }
    sort.Sort(keys)
    // Expected order determined by natural sorting with digit comparison.
    // Explanation:
    // "A1b" comes first (capital 'A' is less than 'a') 
    // then among values starting with 'a', the sorter compares the numeric parts.
    // This expected order is based on the implementation, falling back to comparing digit-run lengths.
    expected := []string{"A1b", "a1B", "a1a", "a1b", "a01b", "a001b"}
    for i, v := range keys {
        if v.String() != expected[i] {
            t.Errorf("Expected %s at index %d, got %s", expected[i], i, v.String())
        }
    }
}

// TestKeyListChannel tests sorting of non-numeric, non-string types (such as channels and slices)
// and verifies that their sort order falls back to reflect.Kind ordering.
func TestKeyListChannel(t *testing.T) {
    // Test with two channels where both are of kind reflect.Chan:
    ch1 := make(chan int)
    ch2 := make(chan int)
    ks := keyList{
        reflect.ValueOf(ch1),
        reflect.ValueOf(ch2),
    }
    sort.Sort(ks)
    // Since both values have the same Kind, ordering is not changed by sort.Sort.
    // We simply verify that the order remains one of the original orders.
    firstChan := ks[0].Interface().(chan int)
    if firstChan != ch1 && firstChan != ch2 {
        t.Errorf("Unexpected ordering for channels")
    }

    // Test with a mix of a slice and a channel (non-numeric, non-string types)
    s := []int{1, 2, 3}
    ks = keyList{
        reflect.ValueOf(s),
        reflect.ValueOf(ch1),
    }
    sort.Sort(ks)
    // By reflect.Kind ordering, we expect that if chan’s Kind is less than slice’s Kind,
    // then the channel appears first.
    if ks[0].Kind() != reflect.Chan {
        t.Errorf("Expected channel to appear first due to kind ordering, got %v", ks[0].Kind())
    }
}
// TestKeyListFuncAndSlice tests sorting when a function and a slice are compared.
func TestKeyListFuncAndSlice(t *testing.T) {
    fn := func() {}
    slice := []int{1, 2, 3}
    ks := keyList{
        reflect.ValueOf(slice),
        reflect.ValueOf(fn),
    }

    sort.Sort(ks)

    if ks[0].Kind() != reflect.Func {
        t.Errorf("Expected function to come first, got kind: %v", ks[0].Kind())
    }
}

// TestKeyListInvalid tests sorting when keyList contains an invalid reflect.Value.
func TestKeyListInvalid(t *testing.T) {
    invalidVal := reflect.Value{}
    ks := keyList{
        invalidVal,
        reflect.ValueOf(42), // int has kind reflect.Int (2)
    }

    sort.Sort(ks)

    if ks[0].Kind() != reflect.Invalid {
        t.Errorf("Expected first element to be Invalid, got %v", ks[0].Kind())
    }

    // Additionally, verify that sorting multiple invalid values does not fail.
    ks = keyList{
        reflect.Value{},
        reflect.Value{},
    }
    sort.Sort(ks)
    if len(ks) != 2 {
        t.Errorf("Expected 2 elements, got %d", len(ks))
    }
}

func TestKeyListArray(t *testing.T) {
    type intArray [3]int
    a := intArray{1, 2, 3}
    b := intArray{4, 5, 6}
    ks := keyList{
        reflect.ValueOf(a),
        reflect.ValueOf(b),
    }
    sort.Sort(ks)
    // Since both elements are of kind reflect.Array, the order is not changed.
    if !reflect.DeepEqual(ks[0].Interface(), a) {
        t.Errorf("Expected first element to be unchanged, got %v", ks[0].Interface())
    }
}

// TestKeyListCustomDigitRun tests sorting of strings with digit runs that include leading zeros.
func TestKeyListCustomDigitRun(t *testing.T) {
    ks := keyList{
        reflect.ValueOf("a001"),
        reflect.ValueOf("a01"),
    }
    sort.Sort(ks)
    // For these strings, after unwrapping letters the digit-run comparison shows that although both
    // evaluate to the same numeric value, the digit run length (ai) differs and thus "a01" should come first.
    expected := []string{"a01", "a001"}
    for i, v := range ks {
        if v.String() != expected[i] {
            t.Errorf("Index %d: expected %s, got %s", i, expected[i], v.String())
        }
    }
}
// TestKeyListPrefix verifies that strings where one is a prefix of another are sorted correctly.
func TestKeyListPrefix(t *testing.T) {
    keys := keyList{
        reflect.ValueOf("abc"),
        reflect.ValueOf("ab"),
    }
    sort.Sort(keys)
    expected := []string{"ab", "abc"}
    for i, v := range keys {
        if v.String() != expected[i] {
            t.Errorf("At index %d, expected %s, got %s", i, expected[i], v.String())
        }
    }
}

// TestKeyListNaturalNumeric verifies natural ordering for strings with embedded numeric parts.
func TestKeyListNaturalNumeric(t *testing.T) {
    keys := keyList{
        reflect.ValueOf("a9"),
        reflect.ValueOf("a10"),
        reflect.ValueOf("a11"),
        reflect.ValueOf("a8"),
    }
    sort.Sort(keys)
    expected := []string{"a8", "a9", "a10", "a11"}
    for i, v := range keys {
        if v.String() != expected[i] {
            t.Errorf("At index %d, expected %s, got %s", i, expected[i], v.String())
        }
    }
}

// TestKeyListIdenticalValues checks that sorting works correctly when all values are identical.
func TestKeyListIdenticalValues(t *testing.T) {
    keys := keyList{
        reflect.ValueOf("same"),
        reflect.ValueOf("same"),
        reflect.ValueOf("same"),
    }
    sort.Sort(keys)
    expected := []string{"same", "same", "same"}
    for i, v := range keys {
        if v.String() != expected[i] {
            t.Errorf("At index %d, expected %s, got %s", i, expected[i], v.String())
        }
    }
}

// TestKeyListMixedNilAndNonNilInterface tests sorting when the keyList contains a nil interface and a non-nil value.
func TestKeyListMixedNilAndNonNilInterface(t *testing.T) {
    var nilInterface interface{} = nil
    ks := keyList{
        reflect.ValueOf(5),
        reflect.ValueOf(nilInterface),
    }
    sort.Sort(ks)
    // The nil interface should come first because its kind is reflect.Invalid.
    if ks[0].Kind() != reflect.Invalid {
        t.Errorf("Expected first element to have Invalid kind, got %v", ks[0].Kind())
    }
}
// TestKeyListMapAndSlice tests sorting of non-numeric, non-string types (map versus slice).
// This verifies that when keyFloat fails, the sorter falls back to comparing reflect.Kind values.
func TestKeyListMapAndSlice(t *testing.T) {
    // Create a map and a slice.
    m := map[int]int{1: 10}
    s := []int{1, 2, 3}
    ks := keyList{
        reflect.ValueOf(s),
        reflect.ValueOf(m),
    }
    sort.Sort(ks)
    // According to reflect.Kind ordering, reflect.Map should come before reflect.Slice.
    if ks[0].Kind() != reflect.Map {
        t.Errorf("Expected first element to be a map, got %v", ks[0].Kind())
    }
}

// TestKeyListDeepInterface tests sorting of values wrapped in several layers of interfaces and pointers.
// It verifies that deep unwrapping correctly retrieves the underlying numeric value.
func TestKeyListDeepInterface(t *testing.T) {
    // Create an underlying int value.
    a := 42
    // Wrap it in an interface.
    i1 := interface{}(a)
    // Wrap that in a pointer.
    p1 := &i1
    // Wrap the pointer in an interface.
    i2 := interface{}(p1)
    // And wrap that in a pointer.
    p2 := &i2
    // Finally, wrap that in one last interface.
    i3 := interface{}(p2)

    // Now create a keyList with i3 and another integer value.
    ks := keyList{
        reflect.ValueOf(i3),
        reflect.ValueOf(10),
    }
    sort.Sort(ks)

    // After unwrapping, 10 (numeric value 10) should come before 42.
    // We unwrap ks[0] to verify its numeric value.
    v := ks[0]
    for (v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr) && !v.IsNil() {
        v = v.Elem()
    }
    if int(v.Int()) != 10 {
        t.Errorf("Expected first element to be 10 after unwrapping, got %v", int(v.Int()))
    }
}
// TestKeyListStringDigitLetter tests ordering when digits and letters are mixed in string keys.
func TestKeyListStringDigitLetter(t *testing.T) {
    // Create a keyList with strings where at the first non-equal character one is a digit and the other a letter.
    keys := keyList{
        reflect.ValueOf("a1"),
        reflect.ValueOf("aa"),
        reflect.ValueOf("a0"),
        reflect.ValueOf("ab"),
    }
    sort.Sort(keys)

    // Expected order:
    // "a0": at index1 '0'
    // "a1": at index1 '1'
    // "aa": at index1 'a' (letter beats a digit in the fallback branch)
    // "ab": at index1 'b'
    expected := []string{"a0", "a1", "aa", "ab"}
    for i, v := range keys {
        if v.String() != expected[i] {
            t.Errorf("At index %d, expected %s, got %s", i, expected[i], v.String())
        }
    }
}

// TestKeyListDeepNilWrapping tests sorting when a nil value is wrapped in multiple levels of interface and pointer.
func TestKeyListDeepNilWrapping(t *testing.T) {
    // Create a nil int pointer.
    var nilInt *int = nil
    // Wrap the nil pointer in an interface.
    var iface interface{} = nilInt
    // Create a pointer to that interface.
    p1 := &iface
    // Wrap that pointer in an interface.
    iface2 := interface{}(p1)
    // And create another pointer on top.
    p2 := &iface2

    // Create a keyList with a valid numeric value and the deeply wrapped nil.
    ks := keyList{
        reflect.ValueOf(p2),
        reflect.ValueOf(5),
    }
    sort.Sort(ks)

    // After sorting, the non-nil numeric value should come first.
    v := ks[0]
    // Unwrap all interface/pointer layers.
    for (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && !v.IsNil() {
        v = v.Elem()
    }
    if int(v.Int()) != 5 {
        t.Errorf("Expected first element to be numeric 5 after unwrapping, got %v", v)
    }
}