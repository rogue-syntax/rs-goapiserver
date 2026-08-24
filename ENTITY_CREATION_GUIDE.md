# Guide: Creating New Entities

This document outlines the established pattern for creating new data entities in the `rs-goapiserver` project. To ensure consistency, please follow these steps precisely. We will use a "Product" entity as an example.

## Step 1: Create the Entity File

Create a new package and file for your entity inside the `entities/` directory.
-   Example: `entities/product/product.go`

## Step 2: Define the `Internal` and `External` Structs

In your new file, define two structs:

1.  **`[EntityName]Internal`**: This struct represents the full database record, including any sensitive or internal-only fields.
2.  **`[EntityName]External`**: This struct represents the data as it should be exposed to the public API. It should omit any sensitive fields.

```go
// In entities/product/product.go
package product

import (
	"database/sql"
	"github.com/pkg/errors"
	"github.com/rogue-syntax/rs-goapiserver/database"
)

type ProductInternal struct {
	ProductID      int
	SKU            string
	Name           string
	Description    string
	Price          float64
	StockQuantity  int
	InternalNotes  string // Internal-only field
}

type ProductExternal struct {
	ProductID     int
	SKU           string
	Name          string
	Description   string
	Price         float64
	StockQuantity int
}
```

## Step 3: Create "Scanner" Functions

A "scanner" function is responsible for taking `*sql.Rows` from a database query and mapping the columns to your struct fields. You need one for both the `Internal` and `External` versions.

**Important:** The order of variables in `rows.Scan()` must exactly match the column order of the `SELECT` query you will write in the next step.

```go
// In entities/product/product.go

func ScanProductInternal(rows *sql.Rows) (*ProductInternal, error) {
	var p ProductInternal
	if rows.Next() {
		// The order here must match your SELECT statement's columns
		err := rows.Scan(
			&p.ProductID,
			&p.SKU,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.StockQuantity,
			&p.InternalNotes,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return &p, nil
	} else {
		// Return nil, nil if no row was found
		return nil, nil
	}
}

func ScanProductExternal(rows *sql.Rows) (*ProductExternal, error) {
	var p ProductExternal
	if rows.Next() {
		// The order here must match your SELECT statement's columns
		err := rows.Scan(
			&p.ProductID,
			&p.SKU,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.StockQuantity,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return &p, nil
	} else {
		return nil, nil
	}
}
```

## Step 4: Create "Finder" Functions

A "finder" function executes a database query and uses the appropriate scanner to return an instance of your entity. These functions use the global `database.DB` object.

```go
// In entities/product/product.go

func FindProductInternalByID(productID int) (*ProductInternal, error) {
	// Note: "SELECT *" assumes column order matches the struct field order,
	// which is fragile. Listing columns explicitly is safer.
	rows, err := database.DB.Query("SELECT ProductID, SKU, Name, Description, Price, StockQuantity, InternalNotes FROM ProductInternalView WHERE ProductID = ?", productID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	
	// Use the scanner to process the result
	product, err := ScanProductInternal(rows)
	return product, errors.WithStack(err)
}

func FindProductExternalBySKU(sku string) (*ProductExternal, error) {
	rows, err := database.DB.Query("SELECT ProductID, SKU, Name, Description, Price, StockQuantity FROM ProductExternalView WHERE SKU = ?", sku)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	product, err := ScanProductExternal(rows)
	return product, errors.WithStack(err)
}
```

## Step 5: Create the Internal-to-External Conversion Function

Finally, create a utility function to safely convert an `Internal` model to an `External` one before sending it as an API response. This function manually copies the shared fields.

```go
// In entities/product/product.go

func ProductInternalToExternal(pi *ProductInternal, px *ProductExternal) *ProductExternal {
	px.ProductID = pi.ProductID
	px.SKU = pi.SKU
	px.Name = pi.Name
	px.Description = pi.Description
	px.Price = pi.Price
	px.StockQuantity = pi.StockQuantity
	return px
}
```

By following these five steps, you can add new entities that are consistent with the existing design of the application.
