terraform {
  backend "http" {
    # address = "https://2nh515zrbf.execute-api.us-east-1.amazonaws.com/tfstate/test"
    # lock_address = "https://2nh515zrbf.execute-api.us-east-1.amazonaws.com/tfstate/test"
    # unlock_address = "https://2nh515zrbf.execute-api.us-east-1.amazonaws.com/tfstate/test"
    # address = "http://localhost:8081/tfstates/test"
    address = "https://2nh515zrbf.execute-api.us-east-1.amazonaws.com/tfstates/test"
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
