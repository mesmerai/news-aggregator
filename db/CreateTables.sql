CREATE TABLE Source (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255)
);

CREATE TABLE Article (
	id SERIAL PRIMARY KEY, 
  source_id INT references Source(id),
 	author      VARCHAR(255),
	title       VARCHAR(255),
	description VARCHAR(255),
	url         VARCHAR(255),
	url_to_image  VARCHAR(255),
	published_at DATE,
	content     VARCHAR(255)
);