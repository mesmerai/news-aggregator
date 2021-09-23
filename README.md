# News Aggregator <WIP>

There are 2 microservices:   

## ncollector 
Service to fetch articles via News API and populate the DB.

Features:
- fetch news ByCountry (Italy and Australia supported)
- populate sources and domains in the DB
- fetch news Globally given a set of feeds
- .. 

> Feeds can be set as Favourite from the web ui from the user.  
> There's a limit of 50 API requests every 6h as per NEWS API Dev(free) plan
> Therefore, setting up a limited/restricted set of feeds limits the total number of API calls

## visualizer
Service that provides a web interface to search news with different filters:
- add/remove Favourite feeds from the menus on the left side (save config to DB)
- search articles per country: Italy, Australia or Global (search from DB)


# Setup 

## Set News Api Key

Register API Key at https://newsapi.org/ to retrieve News   

Then set the following Environment variable:
```
export NEWS_API_KEY="<news-api-key-here>"
export DB_PASSWORD="<postgres-db-password-here>
```

## Start the Postgres Docker Image
Build the image from the ```db/Dockerfile```:
```
$ sudo docker build -t mesmerai/news-postgres db
```

Run the image passing the POSTGRES_PASSWORD as parameter:

```
$ sudo docker run --name news-postgres -p 5432:5432 -e POSTGRES_PASSWORD="<postgres-db-password-here> -d mesmerai/news-postgres
```

Check running docker image:
```
$ sudo docker ps -all
CONTAINER ID   IMAGE                    COMMAND                  CREATED          STATUS          PORTS                    NAMES
1509e02ade12   mesmerai/news-postgres   "docker-entrypoint.s…"   36 seconds ago   Up 35 seconds   0.0.0.0:5432->5432/tcp   news-postgres
```

Connect to DB and check if Tables are created:
```
# psql -h localhost -p 5432 -U news_db_user -d news -W
Password: 
psql (13.4)
Type "help" for help.

-- list tables
news=# \d
                 List of relations
 Schema |      Name      |   Type   |    Owner     
--------+----------------+----------+--------------
 public | article        | table    | news_db_user
 public | article_id_seq | sequence | news_db_user
 public | source         | table    | news_db_user
 public | source_id_seq  | sequence | news_db_user
(4 rows)

```

### To rebuild the image
```
-- list containers
$ sudo docker ps -all
-- stop container
$ sudo docker stop <container-id>
-- remove container
$ sudo docker rm <container-id>
-- list images
$ sudo docker images
-- remove the image
$ sudo docker rmi mesmerai/news-postgres
```

Then rebuild (see above).   

# Dev Takeaways

## Project init
Create ```go.mod```:
```
go mod init github.com/mesmerai/news-aggregator

```

Add module requirements:
```
$ go mod tidy
```

## Data returned from NEWS API request

Each article is an object within the ```articles``` array. 

```
{
  "status": "ok",
  "totalResults": 958,
  "articles": [
    {
      "source": {
        "id": null,
        "name": "tripwire.com"
      },
      "author": null,
      "title": "Clearing Up Elements of Cloud Security",
      "description": "cloud security,brent,people,tripwire cybersecurity podcast,raymond,yeah,elements of cloud,clearing up elements,saas,things,tim erlin,service",
      "url": "https://www.tripwire.com/state-of-security/podcast/clearing-up-elements-of-cloud-security/",
      "urlToImage": "https://www.tripwire.com/state-of-security/wp-content/uploads/sites/3/Talking-Cybersecurity-800x443-1.png",
      "publishedAt": "2021-09-14T03:23:00Z",
      "content": "In this episode, Tripwire’s Brent Holder and Raymond Kirk discuss what cloud security means today. Breaking down the different aspects of cloud security controls, they cover the technology, security … [+16835 chars]"
    },
    {
      "source": {
        "id": null,
        "name": "Seeking Alpha"
      },
```

Useful tool to generate the Go struct from the JSON returned by the request: [JSON-to-GO](https://mholt.github.io/json-to-go/).   
Our ```struct```:      

```
type AutoGenerated struct {
	Status       string `json:"status"`
	TotalResults int    `json:"totalResults"`
	Articles     []struct {
		Source struct {
			ID   interface{} `json:"id"`
			Name string      `json:"name"`
		} `json:"source"`
		Author      interface{} `json:"author"`
		Title       string      `json:"title"`
		Description string      `json:"description"`
		URL         string      `json:"url"`
		URLToImage  string      `json:"urlToImage"`
		PublishedAt time.Time   `json:"publishedAt"`
		Content     string      `json:"content"`
	} `json:"articles"`
}
```





# Appendix

## Install Docker (Fedora 34)

```
# dnf config-manager \
    --add-repo \
    https://download.docker.com/linux/fedora/docker-ce.repo

# dnf install docker-ce docker-ce-cli containerd.io
```

Start Docker (errors and how to fix)
```
-- error running dockerd
failed to start daemon: Error initializing network controller: list bridge addresses failed: PredefinedLocalScopeDefaultNetworks List: [172.17.0.0/16 172.18.0.0/16 172.19.0.0/16 172.20.0.0/16 172.21.0.0/16 172.22.0.0/16 172.23.0.0/16 172.24.0.0/16 172.25.0.0/16 172.26.0.0/16 172.27.0.0/16 172.28.0.0/16 172.29.0.0/16 172.30.0.0/16 172.31.0.0/16 192.168.0.0/20 192.168.16.0/20 192.168.32.0/20 192.168.48.0/20 192.168.64.0/20 192.168.80.0/20 192.168.96.0/20 192.168.112.0/20 192.168.128.0/20 192.168.144.0/20 192.168.160.0/20 192.168.176.0/20 192.168.192.0/20 192.168.208.0/20 192.168.224.0/20 192.168.240.0/20]: no available network

-- fix
# ip link add name docker0 type bridge
# ip addr add dev docker0 172.17.0.1/16
```

Start Docker
```
# systemctl start docker.service

-- enable at boot
# systemctl enable docker.service
```




## Install Postgres (on Fedora 34)

```
# dnf install postgresql-server

-- init
# /usr/bin/postgresql-setup --initdb

-- enable at boot
# systemctl enable postgresql

-- start
# systemctl start postgresql
```

Login
```
$ sudo su - postgres
$ psql
psql (13.4)
Type "help" for help.

postgres=#
```

List Databases
```
postgres-# \l
```

List tables
```
news=# \d
               List of relations
 Schema |      Name      |   Type   |  Owner   
--------+----------------+----------+----------
 public | article        | table    | postgres
 public | article_id_seq | sequence | postgres
 public | source         | table    | postgres
 public | source_id_seq  | sequence | postgres
(4 rows)

```

Describe table 'Articles'
```
news-# \d articles
                                       Table "public.article"
    Column    |          Type          | Collation | Nullable |               Default               
--------------+------------------------+-----------+----------+-------------------------------------
 id           | integer                |           | not null | nextval('article_id_seq'::regclass)
 source_id    | integer                |           |          | 
 author       | character varying(255) |           |          | 
 title        | character varying(255) |           |          | 
 description  | character varying(255) |           |          | 
 url          | character varying(255) |           |          | 
 url_to_image | character varying(255) |           |          | 
 published_at | date                   |           |          | 
 content      | character varying(255) |           |          | 
Indexes:
    "article_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "article_source_id_fkey" FOREIGN KEY (source_id) REFERENCES source(id)

```

Create user (alias to CREATE ROLE)
```
# CREATE USER db_user WITH PASSWORD '*************';
```
Grant permissions to DB
```
# GRANT ALL PRIVILEGES ON DATABASE news to db_user;
# GRANT CONNECT ON DATABASE news TO db_user;
# GRANT USAGE ON SCHEMA public TO db_user;
```

