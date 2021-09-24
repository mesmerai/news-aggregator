# News Aggregator 

There are 2 microservices:   

## ncollector 
Service to fetch articles via News API and populate the DB.

What it does/provides:
- fetch of news ByCountry (Italy and Australia supported)
- populate sources and domains in the DB
- fetch of news Globally given a set of feeds


> Favourite Feeds can be set from the user via web ui (visualizer).  
> NEWS API Dev(free) plan is limited to 50 API requests every 6h.  
> It's recommended to set up a restricted set of feeds to limit the total number of API calls.

## visualizer
Service to provide a web interface for reading records stored in the db.

What it does/provides:  
- search of articles with/without keyword 
- search articles per country: Italy, Australia or Global  
- add/remove Favourite feeds from the menus on the left side (config saved in the DB)


A that's how it looks like:  
![News Aggregator](./images/news-aggregator.png)

# Setup 

## Environment 

Register API Key at https://newsapi.org/ to retrieve News   

Then set the following Environment variable:
```
export NEWS_API_KEY="<news-api-key-here>"
export DB_PASSWORD="<postgres-db-password-here>
```

The database host variable is required as needs to be set to ```db``` when running docker-compose and to ```localhost``` when running a standalone docker image for postgres (see below).
```
-- with docker-compose
$ export DB_HOST="db"

-- with docker
$ export DB_HOST="localhost"
```


## Start with Docker compose
Build
```
$ sudo docker-compose build --build-arg NEWS_API_KEY=--build-arg NEWS_API_KEY=<news-api-key-here> --build-arg DB_PASSWORD=<db-password-here> 
```
Run
```
$ sudo docker-compose up
```

To troubleshoot *ncollector* startup:
```
2021/09/23 14:10:06 Initiate Connection to DB.
2021/09/23 14:10:06 Error connecting to DB => dial tcp 127.0.0.1:5432: connect: connection refused
```
Solved by specifying the DB_HOST as env parameter and implementing Retries on DB connection.   

Now, to troubleshoot: 
```
Error connecting to DB => dial tcp: lookup local on 127.0.0.11:53: no such host
```




## Start Postgres Docker Image
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

## Start ncollector Docker image

Build the image from ncollector/Dockerfile:

```
$ sudo docker build --build-arg NEWS_API_KEY=<news-api-key-here> --build-arg DB_PASSWORD=<db-password-here> -t mesmerai/ncollector ncollector
```






### Other useful docker and docker-compose commands
```
-- list containers
$ sudo docker ps -all
$ sudo docker container ls -a

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

To troubleshoot:
```
$ sudo docker logsv <mycontainer>
$ docker exec -it <mycontainer> bash
```

Networks
```
$ sudo docker network ls
$ sudo docker network prune
```

docker-compose
```
-- Stops containers and removes containers, networks, volumes, and images created by up
$ sudo docker-compose down
```



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

