provider "google" {
  project     = "fog-computing-428109"
  region      = "europe-west6"
  credentials = file("/opt/fogComputing/fog-computing-428109-0a1de1c2f7f2.json")
}

resource "google_compute_network" "vpc_network" {
  name                    = "fog-computing-network"
  auto_create_subnetworks = false
  mtu                     = 1460
}

resource "google_compute_subnetwork" "default" {
  name          = "fog-computing-subnet"
  ip_cidr_range = "10.0.1.0/24"
  region        = "europe-west6"
  network       = google_compute_network.vpc_network.id
}


resource "google_compute_firewall" "ssh-rule" {
  name = "demo-ssh"
  network = google_compute_network.vpc_network.name
  allow {
    protocol = "tcp"
    ports = ["22"]
  }
  target_tags = ["ssh"]
  source_ranges = ["0.0.0.0/0"]
}

resource "google_compute_firewall" "router-rule" {
  name = "demo-router"
  network = google_compute_network.vpc_network.name
  allow {
    protocol = "tcp"
    ports = ["5001"]
  }
  target_tags = ["router"]
  source_ranges = ["0.0.0.0/0"]
}

# Create a single Compute Engine instance
resource "google_compute_instance" "default" {
  name         = "router"
  machine_type = "e2-standard-2"
  zone         = "europe-west6-a"
  tags         = ["ssh", "router"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    }
  }

  metadata_startup_script = "sudo apt-get update; sudo apt-get install -yq git build-essential python3-pip rsync golang tmux; touch /root/authorized_keys; curl -X GET https://github.com/nkowallik.keys >> /root/.ssh/authorized_keys; curl -X GET https://github.com/numyalai.keys >> /root/.ssh/authorized_keys"

  network_interface {
    subnetwork = google_compute_subnetwork.default.id

    access_config {
      # external IP address
    }
  }
}