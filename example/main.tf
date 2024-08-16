terraform {
  backend "http" {
    address = "https://5arjjg54d2uuen2hp2rsmjxg5y0rkhju.lambda-url.us-east-1.on.aws/tfstates/test"
    username = "kamuri"
    password = "test"
  }
}

provider "local" {
  
}

resource "local_file" "test" {
  filename = "${path.module}/created.example"
  content = "this is an example resource"
}
