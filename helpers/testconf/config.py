import os
import yaml
import logging as log


class TestConfig(object):
    def __init__(self):
        self.otg_host = "https://localhost"
        self.otg_ports = ["localhost:5555", "localhost:5555"]

        self.load()

    def load(self, path="test-config.yaml"):
        if os.path.exists(path):
            log.info("Loading test configuration from %s" % path)
            with open(path, "r") as yml:
                yml_dict = yaml.safe_load(yml)

                if "otg_host" in yml_dict:
                    self.otg_host = yml_dict["otg_host"]
                if "otg_ports" in yml_dict:
                    self.otg_ports = yml_dict["otg_ports"]
        else:
            log.info("Using default test configuration")
