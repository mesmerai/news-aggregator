# ref. https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/sql_database_instance

resource "google_sql_database_instance" "news-db" {
    name = var.db_name
    database_version = var.db_version
    region =  var.db_region
    deletion_protection = false

    settings {
        tier = var.db_tier
        disk_size = var.db_disk_size
    }
}

resource "google_sql_user" "news-db-user" {
    count = 1
    instance = google_sql_database_instance.news-db.name
    name = var.db_user
    password = var.db_password
}


