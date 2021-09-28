variable "db_tier" { default = "db-f1-micro"} //cheaper
variable "db_name" { default = "news"}
variable "db_region" { default = "us-west1" }
variable "db_disk_size" { default = "12"}
variable "db_version" { default = "POSTGRES_13"}

variable "db_user" { default = "news_db_user" }

// set as input parameter
variable "db_password" {}




