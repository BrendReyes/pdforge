# rmpage Command - Test Coverage Summary

## Test Results
✅ **All 34 tests passed**

## Valid Input Test Cases (Passing)

### Basic Cases
- Single page: `8` → `[8]`
- Range of pages: `1-18` → `[1, 2, ..., 18]`
- Combination of single and range: `1,6-11,17` → `[1, 6, 7, 8, 9, 10, 11, 17]`
- Reverse order input: `17,1,6-11` → `[1, 6, 7, 8, 9, 10, 11, 17]` (auto-sorted)
- All duplicates: `5,5,5` → `[5]` (deduplicated)

### Whitespace Handling
- Spaces around commas: `1 , 5 , 10` → `[1, 5, 10]`
- Spaces around range: `1 - 5` → `[1, 2, 3, 4, 5]`
- Tabs and spaces mixed: `1,  2  -  4  , 7` → `[1, 2, 3, 4, 7]`

### Single Element Ranges
- Single element range: `5-5` → `[5]`
- Single element range with spaces: `10 - 10` → `[10]`

### Large Numbers
- Large page numbers: `1000-1005` → `[1000, 1001, 1002, 1003, 1004, 1005]`
- Mix of small and large: `1,5,1000,1002-1004` → `[1, 5, 1000, 1002, 1003, 1004]`

### Complex Combinations
- Complex multi-part: `1,3-5,7,10-15,20` → `[1, 3, 4, 5, 7, 10, 11, 12, 13, 14, 15, 20]`
- Overlapping ranges: `1-5,3-8` → `[1, 2, 3, 4, 5, 6, 7, 8]`
- Completely overlapping: `1-10,1-10` → `[1, 2, ..., 10]`

## Invalid Input Test Cases (Properly Rejected)

### Format Errors
- Multiple dashes in range: `1-5-10` ❌
- Non-numeric input: `abc` ❌
- Inverted range: `10-5` ❌

### Invalid Page Numbers
- Negative numbers: `-5` ❌
- Negative in range: `-5-10` ❌
- Zero page number: `0` ❌
- Zero in range: `0-5` ❌

### Empty/Malformed Input
- Empty string: `` ❌
- Only comma: `,` ❌
- Only dashes: `-` ❌
- Leading comma: `,1,2` ❌
- Trailing comma: `1,2,` ❌
- Consecutive commas: `1,,2` ❌

### Type Errors
- Mixed valid and invalid: `1,abc,5` ❌
- Floating point: `1.5,2.3` ❌
- Hexadecimal: `0xFF` ❌
- Text mixed with numbers: `page5` ❌
- Special characters: `1,2@5` ❌

### Range Validation
- Very large range that inverts: `9999-1` ❌

## Implementation Features

✅ Automatic deduplication (map-based storage)
✅ Automatic sorting (bubble sort)
✅ Comprehensive whitespace handling
✅ Range validation (no inverted ranges)
✅ Page number validation (must be positive)
✅ Clear error messages for all failure cases
✅ Handles overlapping ranges correctly
✅ Supports complex multi-part specifications

## Usage Examples

```bash
# Single page
pdforge rmpage input.pdf 8

# Range
pdforge rmpage input.pdf 1-3

# Combination
pdforge rmpage input.pdf 1,6-11,17
```
