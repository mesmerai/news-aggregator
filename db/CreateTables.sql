CREATE TABLE Sources (
		id SERIAL PRIMARY KEY,
		name TEXT
);

CREATE TABLE Domains (
		id SERIAL PRIMARY KEY,
		name TEXT
);

CREATE TABLE Articles (
	id SERIAL PRIMARY KEY, 
  source_id INT references Sources(id),
  domain_id INT references Domains(id),
 	author      TEXT,
	title       TEXT,
	description TEXT,
	url         TEXT,
	url_to_image  TEXT,
	published_at TIMESTAMP with time zone,
	content     TEXT,
	country TEXT,
	language TEXT,
	category TEXT 
);

