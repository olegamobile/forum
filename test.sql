CREATE TABLE IF NOT EXISTS posts_categories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	post_id INTEGER NOT NULL,
	category_id INTEGER NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (post_id) REFERENCES posts(id),
	FOREIGN KEY (category_id) REFERENCES categories(id),
	UNIQUE (post_id, category_id) 
);


INSERT OR IGNORE INTO categories (name) VALUES ('test');