package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the SQLite3 database with the firearms table
func InitDB(dbPath string) (*sql.DB, error) {
	// Open SQLite3 database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create firearms table and indexes
	createTableQuery := `
	-- Creating firearms table with specified and additional relevant columns
	CREATE TABLE IF NOT EXISTS firearms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		brand TEXT NOT NULL,
		name TEXT NOT NULL,
		caliber TEXT NOT NULL,
		type TEXT NOT NULL,
		magazine_capacity INTEGER NOT NULL,
		effective_range INTEGER NOT NULL,
		year INTEGER NOT NULL,
		price INTEGER NOT NULL,
		manufacturer TEXT,
		weight REAL,
		barrel_length REAL,
		action TEXT,
		country_of_origin TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(brand, name)
	);

	-- Creating index on ID for faster lookups
	CREATE INDEX IF NOT EXISTS idx_firearms_id ON firearms(id);

	-- Creating index on brand and name for common queries
	CREATE INDEX IF NOT EXISTS idx_firearms_brand_name ON firearms(brand, name);
	`

	// migration code, only used if the table already exists
	// createTableDuplicateQuery := `
	// 	-- Check for duplicates
	// 	SELECT brand, name, COUNT(*) as count 
	// 	FROM firearms 
	// 	GROUP BY brand, name 
	// 	HAVING count > 1;

	// 	-- Rename existing table
	// 	ALTER TABLE firearms RENAME TO firearms_old;

	// 	-- Create new table with UNIQUE(brand, name)
	// 	CREATE TABLE firearms (
	// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
	// 		brand TEXT NOT NULL,
	// 		name TEXT NOT NULL,
	// 		caliber TEXT NOT NULL,
	// 		type TEXT NOT NULL,
	// 		magazine_capacity INTEGER NOT NULL,
	// 		effective_range INTEGER NOT NULL,
	// 		year INTEGER NOT NULL,
	// 		price INTEGER NOT NULL,
	// 		manufacturer TEXT,
	// 		weight REAL,
	// 		barrel_length REAL,
	// 		action TEXT,
	// 		country_of_origin TEXT,
	// 		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	// 		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	// 		UNIQUE(brand, name)
	// 	);

	// 	-- Migrate data, ignoring duplicates
	// 	INSERT OR IGNORE INTO firearms (
	// 		id, brand, name, caliber, type, magazine_capacity, effective_range, 
	// 		year, price, manufacturer, weight, barrel_length, action, country_of_origin, 
	// 		created_at, updated_at
	// 	) SELECT 
	// 		id, brand, name, caliber, type, magazine_capacity, effective_range, 
	// 		year, price, manufacturer, weight, barrel_length, action, country_of_origin, 
	// 		created_at, updated_at 
	// 	FROM firearms_old;

	// 	-- Drop old table
	// 	DROP TABLE firearms_old;
	// ` 

	// Execute table creation query
	_, err = db.Exec(createTableQuery)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create firearms table: %w", err)
	}

	return db, nil
}


// InsertFirearms adds predefined firearms to the firearms table
func InsertFirearms(db *sql.DB) error {
	// Define the firearms data
	firearms := []struct {
		brand           string
		name            string
		caliber         string
		type_           string
		magazineCap     int
		effectiveRange  int
		year            int
		price           int
		manufacturer    string
		weight          float64
		barrelLength    float64
		action          string
		countryOfOrigin string
	}{
		{"Glock", "19", "9mm Parabellum", "Pistol", 15, 50, 1988, 550, "Glock GmbH", 0.67, 10.2, "Semi-Auto", "Austria"},
		{"Glock", "20", "10mm Auto", "Pistol", 15, 50, 1991, 620, "Glock GmbH", 0.79, 11.7, "Semi-Auto", "Austria"},
		{"Glock", "21", ".45 ACP", "Pistol", 13, 50, 1990, 600, "Glock GmbH", 0.83, 11.7, "Semi-Auto", "Austria"},
		{"H&K", "MP7", "4.6x30mm", "Submachine Gun", 20, 200, 2001, 1700, "Heckler & Koch", 1.9, 18.0, "Select-Fire", "Germany"},
		{"H&K", "MP5", "9mm Parabellum", "Submachine Gun", 30, 200, 1966, 2000, "Heckler & Koch", 2.5, 22.5, "Select-Fire", "Germany"},
		{"H&K", "UMP", ".45 ACP", "Submachine Gun", 25, 100, 1999, 1800, "Heckler & Koch", 2.3, 20.0, "Select-Fire", "Germany"},
		{"H&K", "G36", "5.56x45mm NATO", "Rifle", 30, 600, 1997, 2500, "Heckler & Koch", 3.6, 48.0, "Select-Fire", "Germany"},
		{"H&K", "HK416", "5.56x45mm NATO", "Rifle", 30, 600, 2004, 2700, "Heckler & Koch", 3.4, 36.8, "Select-Fire", "Germany"},
		{"SIG Sauer", "P220", ".45 ACP", "Pistol", 8, 50, 1975, 700, "SIG Sauer", 0.86, 11.2, "Semi-Auto", "Switzerland"},
		{"SIG Sauer", "P226", "9mm Parabellum", "Pistol", 15, 50, 1984, 750, "SIG Sauer", 0.96, 11.2, "Semi-Auto", "Switzerland"},
		{"SIG Sauer", "MPX", "9mm Parabellum", "Submachine Gun", 30, 100, 2013, 1900, "SIG Sauer", 2.7, 20.3, "Select-Fire", "United States"},
		{"Kriss", "Vector", ".45 ACP", "Submachine Gun", 25, 100, 2009, 2200, "Kriss USA", 2.7, 14.0, "Select-Fire", "United States"},
		{"Colt", "1911", ".45 ACP", "Pistol", 7, 50, 1911, 900, "Colt Manufacturing", 1.1, 12.7, "Semi-Auto", "United States"},
		{"Colt", "Anaconda", ".44 Magnum", "Revolver", 6, 100, 1990, 1200, "Colt Manufacturing", 1.5, 15.2, "Double-Action", "United States"},
		{"Colt", "Python", ".357 Magnum", "Revolver", 6, 100, 1955, 1300, "Colt Manufacturing", 1.2, 10.2, "Double-Action", "United States"},
		{"Colt", "AR-15", "5.56x45mm NATO", "Rifle", 30, 600, 1964, 1000, "Colt Manufacturing", 3.2, 50.8, "Semi-Auto", "United States"},
		{"ArmaLite", "AR-19", "9mm Parabellum", "Rifle", 32, 200, 2020, 1500, "ArmaLite", 3.0, 40.6, "Semi-Auto", "United States"},
		{"Kalashnikov", "AK-47", "7.62x39mm", "Rifle", 30, 350, 1949, 800, "Kalashnikov Concern", 4.3, 41.5, "Select-Fire", "Russia"},
		{"Kalashnikov", "AKM", "7.62x39mm", "Rifle", 30, 350, 1959, 850, "Kalashnikov Concern", 3.1, 41.5, "Select-Fire", "Russia"},
		{"Smith & Wesson", "M&P Shield", "9mm Parabellum", "Pistol", 8, 50, 2012, 600, "Smith & Wesson", 0.58, 7.9, "Semi-Auto", "United States"},
		{"Smith & Wesson", "Model 686", ".357 Magnum", "Revolver", 6, 100, 1980, 1000, "Smith & Wesson", 1.3, 10.2, "Double-Action", "United States"},
		{"Smith & Wesson", "M&P15", "5.56x45mm NATO", "Rifle", 30, 600, 2006, 1200, "Smith & Wesson", 3.2, 40.6, "Semi-Auto", "United States"},
		{"Springfield", "Enhanced 1911", ".45 ACP", "Pistol", 7, 50, 1985, 1100, "Springfield Armory", 1.1, 12.7, "Semi-Auto", "United States"},
		{"Springfield", "M1A", "7.62x51mm NATO", "Rifle", 20, 800, 1974, 1800, "Springfield Armory", 4.2, 55.9, "Semi-Auto", "United States"},
		{"Springfield", "XD-M", "9mm Parabellum", "Pistol", 19, 50, 2008, 700, "Springfield Armory", 0.88, 11.7, "Semi-Auto", "United States"},
		{"Beretta", "92FS", "9mm Parabellum", "Pistol", 15, 50, 1976, 800, "Beretta", 0.95, 12.5, "Semi-Auto", "Italy"},
		{"Beretta", "M9A4", "9mm Parabellum", "Pistol", 17, 50, 2021, 900, "Beretta", 0.94, 12.5, "Semi-Auto", "Italy"},
		{"Beretta", "APX", "9mm Parabellum", "Pistol", 17, 50, 2017, 650, "Beretta", 0.80, 10.8, "Semi-Auto", "Italy"},
		{"Benelli", "M4", "12 Gauge", "Shotgun", 7, 50, 1998, 1600, "Benelli Armi", 3.8, 47.0, "Semi-Auto", "Italy"},
		{"Benelli", "Super Black Eagle 3", "12 Gauge", "Shotgun", 4, 50, 2017, 2000, "Benelli Armi", 3.3, 71.1, "Semi-Auto", "Italy"},
		{"Benelli", "Nova", "12 Gauge", "Shotgun", 4, 50, 1999, 900, "Benelli Armi", 3.6, 66.0, "Pump-Action", "Italy"},
		{"KBP", "PP-2000", "9mm Parabellum", "Submachine Gun", 20, 100, 2006, 1400, "KBP Instrument Design Bureau", 1.4, 18.2, "Select-Fire", "Russia"},
		{"Izhmash", "PP-19 Bizon", "9x18mm Makarov", "Submachine Gun", 64, 100, 1996, 1500, "Izhmash", 2.1, 22.5, "Select-Fire", "Russia"},
		{"Nagant", "M1895", "7.62x38mmR", "Revolver", 7, 50, 1895, 500, "Tula Arsenal", 0.8, 11.4, "Double-Action", "Russia"},
		{"Izhmash", "PP-19-01 Vityaz-SN", "9mm Parabellum", "Submachine Gun", 30, 200, 2004, 1600, "Izhmash", 2.9, 23.7, "Select-Fire", "Russia"},
		{"Kalashnikov", "PPK-20", "9mm Parabellum", "Submachine Gun", 30, 200, 2020, 1800, "Kalashnikov Concern", 2.7, 23.7, "Select-Fire", "Russia"},
		{"Kalashnikov", "Saiga-9", "9mm Parabellum", "Carbine", 10, 200, 2010, 1200, "Kalashnikov Concern", 3.2, 34.5, "Semi-Auto", "Russia"},
		{"Yarygin", "MP-443 Grach", "9mm Parabellum", "Pistol", 17, 50, 2003, 600, "Izhevsk Mechanical Plant", 0.95, 11.2, "Semi-Auto", "Russia"},
		{"Izhmash", "Makarov PM", "9x18mm Makarov", "Pistol", 8, 50, 1951, 400, "Izhmash", 0.73, 9.3, "Semi-Auto", "Russia"},
		{"Izhmash", "PSM", "5.45x18mm", "Pistol", 8, 50, 1973, 450, "Izhmash", 0.46, 8.5, "Semi-Auto", "Russia"},
		{"FN", "Five-seveN", "5.7x28mm", "Pistol", 20, 50, 2000, 1100, "FN Herstal", 0.62, 12.2, "Semi-Auto", "Belgium"},
		{"FN", "SCAR-L", "5.56x45mm NATO", "Rifle", 30, 600, 2009, 2500, "FN Herstal", 3.3, 35.1, "Select-Fire", "Belgium"},
		{"FN", "P90", "5.7x28mm", "Submachine Gun", 50, 200, 1990, 2000, "FN Herstal", 2.6, 26.3, "Select-Fire", "Belgium"},
		{"FN", "FAL", "7.62x51mm NATO", "Rifle", 20, 800, 1953, 1500, "FN Herstal", 4.3, 53.3, "Select-Fire", "Belgium"},
		{"Kalashnikov", "Saiga-12", "12 Gauge", "Shotgun", 8, 50, 1997, 1000, "Kalashnikov Concern", 3.6, 43.0, "Semi-Auto", "Russia"},
		{"Kalashnikov", "Saiga-410", ".410 Bore", "Shotgun", 8, 50, 1997, 900, "Kalashnikov Concern", 3.4, 43.0, "Semi-Auto", "Russia"},
		{"Kalashnikov", "Saiga-20", "20 Gauge", "Shotgun", 8, 50, 1997, 950, "Kalashnikov Concern", 3.5, 43.0, "Semi-Auto", "Russia"},
		{"Molot", "Vepr-12", "12 Gauge", "Shotgun", 8, 50, 2003, 1100, "Molot-Oruzhie", 3.9, 43.0, "Semi-Auto", "Russia"},
		{"Kalashnikov", "Saiga-12", "12 Gauge", "Shotgun", 8, 50, 1997, 1000, "Kalashnikov Concern", 3.6, 43.0, "Semi-Auto", "Russia"},
		{"Degtyarev", "RPG-7", "40mm Rocket", "Rocket Launcher", 1, 300, 1961, 2500, "Bazalt", 7.0, 95.0, "Single-Shot", "Russia"},
		{"Raytheon", "FGM-148 Javelin", "127mm Missile", "Missile Launcher", 1, 2500, 1996, 25000, "Raytheon/Lockheed Martin", 22.3, 110.0, "Single-Shot", "United States"},
		{"Lockheed Martin", "Predator SRAW", "140mm Missile", "Missile Launcher", 1, 600, 2002, 15000, "Lockheed Martin", 9.8, 100.0, "Single-Shot", "United States"},
		{"Saab", "AT4", "84mm Rocket", "Rocket Launcher", 1, 300, 1987, 2000, "Saab Bofors Dynamics", 6.7, 100.0, "Single-Shot", "Sweden"},
		{"Colt", "M4 Carbine", "5.56x45mm NATO", "Rifle", 30, 500, 1994, 2000, "Colt Manufacturing", 2.9, 36.8, "Select-Fire", "United States"},
		{"Tula", "PPSh-41", "7.62x25mm Tokarev", "Submachine Gun", 71, 200, 1941, 600, "Tula Arsenal", 3.6, 26.9, "Select-Fire", "Russia"},
		{"Erma", "MP40", "9mm Parabellum", "Submachine Gun", 32, 100, 1940, 700, "Erma Werke", 4.0, 25.1, "Select-Fire", "Germany"},
		{"Mauser", "MG42", "7.92x57mm Mauser", "Machine Gun", 250, 1000, 1942, 3000, "Mauser Werke", 11.6, 53.0, "Full-Auto", "Germany"},
		{"Browning", "M1919", "7.62x51mm NATO", "Machine Gun", 250, 1000, 1919, 2500, "Browning Arms", 14.0, 61.0, "Full-Auto", "United States"},
		{"General Electric", "M134D Minigun", "7.62x51mm NATO", "Rotary Machine Gun", 4000, 1000, 1960, 50000, "General Electric", 38.0, 55.9, "Full-Auto", "United States"},
		{"Ruger", "10/22", ".22 LR", "Rifle", 10, 100, 1964, 300, "Sturm, Ruger & Co.", 2.3, 47.0, "Semi-Auto", "United States"},
		{"Ruger", "Mini-14", "5.56x45mm NATO", "Rifle", 20, 400, 1973, 900, "Sturm, Ruger & Co.", 2.9, 47.0, "Semi-Auto", "United States"},
		{"Remington", "870", "12 Gauge", "Shotgun", 7, 50, 1950, 500, "Remington Arms", 3.6, 71.1, "Pump-Action", "United States"},
		{"Remington", "700", ".308 Winchester", "Rifle", 4, 800, 1962, 800, "Remington Arms", 3.4, 61.0, "Bolt-Action", "United States"},
		{"Winchester", "Model 70", ".30-06 Springfield", "RifleΙ", 5, 800, 1936, 1000, "Winchester Repeating Arms", 3.6, 61.0, "Bolt-Action", "United States"},
		{"CZ", "CZ 75", "9mm Parabellum", "Pistol", 16, 50, 1975, 700, "Česká zbrojovka", 1.0, 12.0, "Semi-Auto", "Czech Republic"},
		{"IWI", "Tavor X95", "5.56x45mm NATO", "Rifle", 30, 500, 2009, 2200, "Israel Weapon Industries", 3.3, 33.0, "Select-Fire", "Israel"},
		{"Steyr", "AUG", "5.56x45mm NATO", "Rifle", 30, 600, 1977, 2100, "Steyr Mannlicher", 3.6, 50.8, "Select-Fire", "Austria"},
		{"Mossberg", "500", "12 Gauge", "Shotgun", 6, 50, 1960, 450, "O.F. Mossberg & Sons", 3.4, 71.1, "Pump-Action", "United States"},
		{"Magnum Research", "Desert Eagle", ".50 AE", "Pistol", 7, 50, 1983, 1500, "Magnum Research", 2.0, 15.2, "Semi-Auto", "United States"},
		{"Beretta", "93R", "9mm Parabellum", "Pistol", 20, 50, 1979, 1200, "Beretta", 1.2, 12.5, "Select-Fire", "Italy"},
		{"H&K", "USP", "9mm Parabellum", "Pistol", 15, 50, 1993, 800, "Heckler & Koch", 0.79, 10.8, "Semi-Auto", "Germany"},
		{"SIG Sauer", "P250", "9mm Parabellum", "Pistol", 17, 50, 2007, 650, "SIG Sauer", 0.82, 10.8, "Semi-Auto", "United States"},
		{"GIAT", "FAMAS", "5.56x45mm NATO", "Rifle", 25, 450, 1978, 2200, "Nexter Systems", 3.6, 48.8, "Select-Fire", "France"},
		{"Accuracy International", "AWP", "7.62x51mm NATO", "Sniper Rifle", 10, 800, 1997, 3000, "Accuracy International", 6.5, 61.0, "Bolt-Action", "United Kingdom"},
		{"SIG Sauer", "SG553", "5.56x45mm NATO", "Rifle", 30, 400, 2009, 2300, "Swiss Arms", 3.2, 34.7, "Select-Fire", "Switzerland"},
	}

	// SQL INSERT query template
	query := `
		INSERT OR IGNORE INTO firearms (
			brand, name, caliber, type, magazine_capacity, effective_range, 
			year, price, manufacturer, weight, barrel_length, action, country_of_origin
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Execute INSERT for each firearm
	for _, gun := range firearms {
		_, err := db.Exec(query,
			gun.brand, gun.name, gun.caliber, gun.type_, gun.magazineCap,
			gun.effectiveRange, gun.year, gun.price, gun.manufacturer, gun.weight,
			gun.barrelLength, gun.action, gun.countryOfOrigin,
		)
		if err != nil {
			return fmt.Errorf("failed to insert firearm %s %s: %w", gun.brand, gun.name, err)
		}
	}

	return nil
}

// Firearm represents the structure of a firearm record
type Firearm struct {
	ID              int     `json:"id"`
	Brand           string  `json:"brand"`
	Name            string  `json:"name"`
	Caliber         string  `json:"caliber"`
	Type            string  `json:"type"`
	MagazineCapacity int     `json:"magazine_capacity"`
	EffectiveRange  int     `json:"effective_range"`
	Year            int     `json:"year"`
	Price           int     `json:"price"`
	Manufacturer    string  `json:"manufacturer"`
	Weight          float64 `json:"weight"`
	BarrelLength    float64 `json:"barrel_length"`
	Action          string  `json:"action"`
	CountryOfOrigin string  `json:"country_of_origin"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// GetFirearmsByBrand retrieves firearms by brand
func GetFirearmsByBrand(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		brand := c.Param("brand")
		if brand == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "brand parameter is required"})
			return
		}

		// Use parameterized query to prevent SQL injection
		rows, err := db.Query("SELECT * FROM firearms WHERE brand = ?", strings.Title(brand))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query database: %v", err)})
			return
		}
		defer rows.Close()

		var firearms []Firearm
		for rows.Next() {
			var f Firearm
			err := rows.Scan(&f.ID, &f.Brand, &f.Name, &f.Caliber, &f.Type, &f.MagazineCapacity,
				&f.EffectiveRange, &f.Year, &f.Price, &f.Manufacturer, &f.Weight, &f.BarrelLength,
				&f.Action, &f.CountryOfOrigin, &f.CreatedAt, &f.UpdatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to scan row: %v", err)})
				return
			}
			firearms = append(firearms, f)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error iterating rows: %v", err)})
			return
		}

		if len(firearms) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("no firearms found for brand: %s", brand)})
			return
		}

		c.JSON(http.StatusOK, firearms)
	}
}

func GetFirearmsByName(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name parameter is required"})
			return
		}

		// Use parameterized query to prevent SQL injection
		rows, err := db.Query("SELECT * FROM firearms WHERE name LIKE '%' || ? || '%' COLLATE NOCASE", strings.Title(name))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query database: %v", err)})
			return
		}
		defer rows.Close()

		var firearms []Firearm
		for rows.Next() {
			var f Firearm
			err := rows.Scan(&f.ID, &f.Brand, &f.Name, &f.Caliber, &f.Type, &f.MagazineCapacity,
				&f.EffectiveRange, &f.Year, &f.Price, &f.Manufacturer, &f.Weight, &f.BarrelLength,
				&f.Action, &f.CountryOfOrigin, &f.CreatedAt, &f.UpdatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to scan row: %v", err)})
				return
			}
			firearms = append(firearms, f)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error iterating rows: %v", err)})
			return
		}

		if len(firearms) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("no firearms found with name: %s", name)})
			return
		}

		c.JSON(http.StatusOK, firearms)
	}
}

// GetFirearmsByCaliber retrieves firearms by partial caliber match
func GetFirearmsByCaliber(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		caliber := c.Param("caliber")
		if caliber == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "caliber parameter is required"})
			return
		}

		// Use LIKE for partial, case-insensitive matching
		rows, err := db.Query("SELECT * FROM firearms WHERE caliber LIKE '%' || ? || '%' COLLATE NOCASE", caliber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query database: %v", err)})
			return
		}
		defer rows.Close()

		var firearms []Firearm
		for rows.Next() {
			var f Firearm
			err := rows.Scan(&f.ID, &f.Brand, &f.Name, &f.Caliber, &f.Type, &f.MagazineCapacity,
				&f.EffectiveRange, &f.Year, &f.Price, &f.Manufacturer, &f.Weight, &f.BarrelLength,
				&f.Action, &f.CountryOfOrigin, &f.CreatedAt, &f.UpdatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to scan row: %v", err)})
				return
			}
			firearms = append(firearms, f)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error iterating rows: %v", err)})
			return
		}

		if len(firearms) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("no firearms found for caliber: %s", caliber)})
			return
		}

		c.JSON(http.StatusOK, firearms)
	}
}

func GetFirearmsByPrice(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		minPriceStr := c.Param("min")
		maxPriceStr := c.Param("max")

		if minPriceStr == "" || maxPriceStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "min and max price parameters are required"})
			return
		}

		// Convert string parameters to integers
		minPrice, err := strconv.Atoi(minPriceStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "min price must be a valid integer"})
			return
		}
		maxPrice, err := strconv.Atoi(maxPriceStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "max price must be a valid integer"})
			return
		}

		if minPrice > maxPrice {
			c.JSON(http.StatusBadRequest, gin.H{"error": "min price cannot be greater than max price"})
			return
		}

		// Query using BETWEEN for price range
		rows, err := db.Query("SELECT * FROM firearms WHERE price BETWEEN ? AND ?", minPrice, maxPrice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query database: %v", err)})
			return
		}
		defer rows.Close()

		var firearms []Firearm
		for rows.Next() {
			var f Firearm
			err := rows.Scan(&f.ID, &f.Brand, &f.Name, &f.Caliber, &f.Type, &f.MagazineCapacity,
				&f.EffectiveRange, &f.Year, &f.Price, &f.Manufacturer, &f.Weight, &f.BarrelLength,
				&f.Action, &f.CountryOfOrigin, &f.CreatedAt, &f.UpdatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to scan row: %v", err)})
				return
			}
			firearms = append(firearms, f)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error iterating rows: %v", err)})
			return
		}

		if len(firearms) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("no firearms found for price range: %d to %d", minPrice, maxPrice)})
			return
		}

		c.JSON(http.StatusOK, firearms)
	}
}

func GetFirearmsByCountry(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		country := c.Param("country")
		if country == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "country parameter is required"})
			return
		}

		// Use LIKE for partial, case-insensitive matching
		rows, err := db.Query("SELECT * FROM firearms WHERE country_of_origin LIKE '%' || ? || '%' COLLATE NOCASE", strings.Title(country))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query database: %v", err)})
			return
		}
		defer rows.Close()

		var firearms []Firearm
		for rows.Next() {
			var f Firearm
			err := rows.Scan(&f.ID, &f.Brand, &f.Name, &f.Caliber, &f.Type, &f.MagazineCapacity,
				&f.EffectiveRange, &f.Year, &f.Price, &f.Manufacturer, &f.Weight, &f.BarrelLength,
				&f.Action, &f.CountryOfOrigin, &f.CreatedAt, &f.UpdatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to scan row: %v", err)})
				return
			}
			firearms = append(firearms, f)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error iterating rows: %v", err)})
			return
		}

		if len(firearms) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("no firearms found for country: %s", country)})
			return
		}

		c.JSON(http.StatusOK, firearms)
	}
}

func GetFirearmsByYear(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		year := c.Param("year")
		if year == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "year parameter is required"})
			return
		}

		// Use LIKE for partial, case-insensitive matching
		rows, err := db.Query("SELECT * FROM firearms WHERE year = ?", year)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query database: %v", err)})
			return
		}
		defer rows.Close()

		var firearms []Firearm
		for rows.Next() {
			var f Firearm
			err := rows.Scan(&f.ID, &f.Brand, &f.Name, &f.Caliber, &f.Type, &f.MagazineCapacity,
				&f.EffectiveRange, &f.Year, &f.Price, &f.Manufacturer, &f.Weight, &f.BarrelLength,
				&f.Action, &f.CountryOfOrigin, &f.CreatedAt, &f.UpdatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to scan row: %v", err)})
				return
			}
			firearms = append(firearms, f)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error iterating rows: %v", err)})
			return
		}

		if len(firearms) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("no firearms found for year: %s", year)})
			return
		}

		c.JSON(http.StatusOK, firearms)
	}
}

func GetFirearmsByType(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		weptype := c.Param("type")
		if weptype == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type parameter is required"})
			return
		}

		// Use LIKE for partial, case-insensitive matching
		rows, err := db.Query("SELECT * FROM firearms WHERE type LIKE '%' || ? || '%' COLLATE NOCASE", strings.Title(weptype))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query database: %v", err)})
			return
		}
		defer rows.Close()

		var firearms []Firearm
		for rows.Next() {
			var f Firearm
			err := rows.Scan(&f.ID, &f.Brand, &f.Name, &f.Caliber, &f.Type, &f.MagazineCapacity,
				&f.EffectiveRange, &f.Year, &f.Price, &f.Manufacturer, &f.Weight, &f.BarrelLength,
				&f.Action, &f.CountryOfOrigin, &f.CreatedAt, &f.UpdatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to scan row: %v", err)})
				return
			}
			firearms = append(firearms, f)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error iterating rows: %v", err)})
			return
		}

		if len(firearms) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("no firearms found for type: %s", weptype)})
			return
		}

		c.JSON(http.StatusOK, firearms)
	}
}

// GetFirearmByID retrieves a firearm by ID
func GetFirearmByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
			return
		}

		var f Firearm
		err := db.QueryRow("SELECT * FROM firearms WHERE id = ?", id).Scan(
			&f.ID, &f.Brand, &f.Name, &f.Caliber, &f.Type, &f.MagazineCapacity,
			&f.EffectiveRange, &f.Year, &f.Price, &f.Manufacturer, &f.Weight, &f.BarrelLength,
			&f.Action, &f.CountryOfOrigin, &f.CreatedAt, &f.UpdatedAt,
		)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("no firearm found with id: %s", id)})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query database: %v", err)})
			return
		}

		c.JSON(http.StatusOK, f)
	}
}

// GetAllFirearms retrieves all firearms
func GetAllFirearms(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT * FROM firearms")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query database: %v", err)})
			return
		}
		defer rows.Close()

		var firearms []Firearm
		for rows.Next() {
			var f Firearm
			err := rows.Scan(&f.ID, &f.Brand, &f.Name, &f.Caliber, &f.Type, &f.MagazineCapacity,
				&f.EffectiveRange, &f.Year, &f.Price, &f.Manufacturer, &f.Weight, &f.BarrelLength,
				&f.Action, &f.CountryOfOrigin, &f.CreatedAt, &f.UpdatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to scan row: %v", err)})
				return
			}
			firearms = append(firearms, f)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error iterating rows: %v", err)})
			return
		}

		c.JSON(http.StatusOK, firearms)
	}
}
	
func main() {
	db, err := InitDB("gundatabase.db")
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}
	defer db.Close()

	// Uncomment/Comment based on new entries or not
	// err = InsertFirearms(db)
	// if err != nil {
	// 	fmt.Println("Error inserting firearms:", err)
	// 	return
	// }

	r := gin.Default()
	r.LoadHTMLGlob("**/*.html")
	r.Static("/static", "./static")

	r.GET("/brand/:brand", GetFirearmsByBrand(db))
	r.GET("/name/:name", GetFirearmsByName(db))
	r.GET("/caliber/:caliber", GetFirearmsByCaliber(db))
	r.GET("/year/:year", GetFirearmsByYear(db))
	r.GET("/type/:type", GetFirearmsByType(db))
	r.GET("/country/:country", GetFirearmsByCountry(db))
	r.GET("/price/:min/:max", GetFirearmsByPrice(db))
	r.GET("/id/:id", GetFirearmByID(db))
	r.GET("/all", GetAllFirearms(db))

	err = r.Run(":4000")
	if err != nil {
		log.Fatal(err)
	}
}