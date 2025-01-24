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

INSERT OR IGNORE INTO posts_categories (post_id, category_id) VALUES ('25', '4');

DELETE FROM categories WHERE id = '7';
DELETE FROM posts_categories WHERE id = '9';

SELECT c.name AS category FROM posts_categories pc JOIN categories c ON pc.category_id = c.id WHERE post_id = 1;

SELECT DISTINCT p.id, p.author, p.title, p.content, p.created_at FROM posts p JOIN posts_categories pc on pc.post_id = p.id JOIN categories cats ON cats.name IN ('tea', 'coffee');
SELECT DISTINCT p.id, p.author, p.title, p.content, p.created_at FROM posts p JOIN posts_categories pc ON pc.post_id = p.id JOIN categories cats ON cats.id = pc.category_id WHERE cats.name IN ('test');

SELECT DISTINCT p.id, p.author, p.title, p.content, p.created_at FROM posts p JOIN posts_categories pc ON pc.post_id = p.id JOIN categories cats ON cats.id = pc.category_id WHERE cats.name IN (%s) GROUP BY p.id, p.author, p.title, p.content, p.created_at HAVING COUNT(DISTINCT cats.name) = %v;

SELECT id, author, title, content, created_at FROM posts WHERE id = 1;

SELECT DISTINCT p.id, p.author, p.title, p.content, p.created_at FROM posts p JOIN posts_categories pc on pc.post_id = p.id JOIN categories cats ON cats.id = pc.category_id WHERE cats.name IN ('coffee', 'tea');