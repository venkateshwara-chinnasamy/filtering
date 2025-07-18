## 🔧 **JSON Filter & Transform Utility**

A powerful command-line tool that reads a JSON file, applies filtering and transformation rules to keys and values, and writes the result to a new JSON file.

---

### ✅ **Features Overview**

This utility supports:

* **Filtering**: Include or exclude key-value pairs based on criteria such as depth, key length, value type, numeric/string value range, string patterns, etc.
* **Transformation**: Modify keys and values in-place using pattern replacements, string padding/truncation, numeric bounding, masking, conditional replacements, and more.
* **Recursive Traversal**: Operates across nested objects and arrays unless explicitly filtered.
* **Array Handling**: Filter or transform array elements by type or value.
* **Flexible Configuration**: All options are composable — multiple filters and transforms can be combined in a single run.

---

### 📘 **Usage**

```bash
go run filter_json.go [options] input.json output.json
```

---

### 🔍 **Filtering Options**

| Option              | Description                                                                                   |
| ------------------- | --------------------------------------------------------------------------------------------- |
| `-mindepth <n>`     | Include only keys at or below a minimum nesting level (root is depth 1)                       |
| `-maxdepth <n>`     | Include only keys at or above a maximum nesting level                                         |
| `-minkeylen <n>`    | Include keys with name length ≥ n                                                             |
| `-maxkeylen <n>`    | Include keys with name length ≤ n                                                             |
| `-novaltype <type>` | Exclude keys with values of this type (`string`, `number`, `bool`, `null`, `object`, `array`) |
| `-minnum <n>`       | For numeric values, include only if value ≥ n                                                 |
| `-maxnum <n>`       | For numeric values, include only if value ≤ n                                                 |
| `-minstrlen <n>`    | Include string values only if length ≥ n                                                      |
| `-maxstrlen <n>`    | Include string values only if length ≤ n                                                      |
| `-strpattern <p>`   | Include string values matching character classes (`upper`, `lower`, `num`, `sym`)             |
| `-nostrpattern <p>` | Exclude strings matching specified character classes                                          |
| `-ignorecase`       | Makes string pattern filters case-insensitive                                                 |

---

### ✨ **Transformation Options**

| Option                                   | Description                                                             |
| ---------------------------------------- | ----------------------------------------------------------------------- |
| `-replaceval <pattern>:<replacement>`    | Replace string **values** matching the pattern with a given replacement |
| `-replacekey <pattern>:<replacement>`    | Replace **key names** matching a pattern                                |
| `-boundnum <min>:<max>`                  | Clamp numeric values within \[min, max] range                           |
| `-boundstrlen <min>:<max>`               | For strings: pad if too short, truncate if too long                     |
| `-defaultval <type>:<value>`             | Replace `null` or empty values of a given type with a default           |
| `-arrayfilter <type>:<filter>`           | Apply a filter (e.g., `-minnum`) to array elements of a specified type  |
| `-renamekeydepth <depth>:<prefix>`       | Add a prefix to all keys at a specific depth                            |
| `-maskval <key>:<mask>`                  | Mask the value of a specific key with a fixed string                    |
| `-condreplace <condition>:<replacement>` | Replace values conditionally (e.g., `value==null`)                      |

---

### 🧪 **Example Scenarios**

Each command demonstrates a specific use case, combining filters and transformations.

---

#### **1. Filter by key length and exclude nulls**

```bash
go run filter_json.go -minkeylen 4 -novaltype null input.json output.json
```

**Effect**: Retains only keys with ≥4 characters and non-null values.

---

#### **2. Limit numeric values to a specific range**

```bash
go run filter_json.go -boundnum 10:100 input.json output.json
```

**Effect**: Numbers <10 are set to 10, >100 set to 100. Applies to all levels, including arrays.

---

#### **3. Rename specific keys**

```bash
go run filter_json.go -replacekey email:contact -replacekey score:points input.json output.json
```

**Effect**: Replaces "email" with "contact" and "score" with "points" recursively.

---

#### **4. Truncate and pad strings**

```bash
go run filter_json.go -boundstrlen 5:8 input.json output.json
```

**Effect**: Strings shorter than 5 characters are padded; longer than 8 are truncated.

---

#### **5. Default values for nulls and empty strings**

```bash
go run filter_json.go -defaultval null:"N/A" -defaultval string:"EMPTY" input.json output.json
```

**Effect**: Replaces all `null` with `"N/A"` and empty strings with `"EMPTY"`.

---

#### **6. Filter array elements by value**

```bash
go run filter_json.go -arrayfilter number:-minnum 10 input.json output.json
```

**Effect**: Filters out numbers <10 from all arrays.

---

#### **7. Add prefix to keys at a certain depth**

```bash
go run filter_json.go -renamekeydepth 2:sub_ input.json output.json
```

**Effect**: Adds `"sub_"` prefix to keys exactly at depth 2.

---

#### **8. Mask specific keys**

```bash
go run filter_json.go -maskval email:***MASKED*** -maskval Name:MASK input.json output.json
```

**Effect**: Masks values of `email` and `Name` keys.

---

#### **9. Conditional value replacement**

```bash
go run filter_json.go -condreplace "value==\"Alice\"":User -condreplace "value==null":Unknown input.json output.json
```

**Effect**: Replaces `"Alice"` with `"User"` and `null` with `"Unknown"`.

---

#### **10. Combined filter and transformation**

```bash
go run filter_json.go -replaceval num:REDACTED -boundnum 0:100 -defaultval null:0 input.json output.json
```

**Effect**:

* Numeric values are bounded \[0–100]
* String values containing digits are replaced with `"REDACTED"`
* Nulls replaced with `0`

---

### 📝 **Tips**

* Pattern values like `upper`, `num`, `sym` can be combined:
  `-strpattern upper,num`
* Use `-ignorecase` to make pattern checks case-insensitive.
* For `-condreplace`, use valid logical expressions like:
  `value==null`, `value=="some string"`, `value>100`

---

### 🧩 **Final Notes**

* This tool is ideal for **data sanitization**, **privacy masking**, **schema transformation**, or **selective JSON extraction**.
* All operations are **non-destructive** unless explicitly overwritten.
* Nesting is handled **recursively**, so transformations apply to deeply nested structures by default.

---
