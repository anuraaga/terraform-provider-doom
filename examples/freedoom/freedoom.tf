terraform {
  required_providers {
    doom = {
      source = "anuraaga/doom"
    }
  }
}

provider "doom" {
  path = "/Applications/GZDoom.app/Contents/MacOS/gzdoom"
}

resource "doom_session" "freedoom1" {
  wad = "freedoom1"
}


resource "doom_session" "freedoom2" {
  wad = "freedoom2"
}
