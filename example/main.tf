terraform {
  backend "http" {
    address = "https://549f2l2abc.execute-api.us-east-1.amazonaws.com/tfstate/test"
    lock_address = "https://549f2l2abc.execute-api.us-east-1.amazonaws.com/tfstate/test"
    unlock_address = "https://549f2l2abc.execute-api.us-east-1.amazonaws.com/tfstate/test"
  }
}

provider "local" {
  
}

resource "local_file" "test" {
  filename = "${path.module}/created.example"
  content = "this is an example resource"
}