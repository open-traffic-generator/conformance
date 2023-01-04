import os
import yaml
import logging as log


class TestConfig(object):
    def __init__(self):
        self.otg_host = "https://localhost:8443"
        self.otg_ports = ["localhost:5555", "localhost:5555"]
        self.otg_speed = "speed_1_gbps"
        self.otg_capture_check = True
        self.otg_iterations = 100
        self.otg_grpc_transport = False

        self.load()
        self.from_env()

    def load(self, path="test-config.yaml"):
        if os.path.exists(path):
            log.info("Loading test configuration from %s" % path)
            with open(path, "r") as yml:
                yml_dict = yaml.safe_load(yml)

                if "otg_host" in yml_dict:
                    self.otg_host = yml_dict["otg_host"]
                if "otg_ports" in yml_dict:
                    self.otg_ports = yml_dict["otg_ports"]
                if "otg_speed" in yml_dict:
                    self.otg_speed = yml_dict["otg_speed"]
                if "otg_capture_check" in yml_dict:
                    self.otg_capture_check = yml_dict["otg_capture_check"]
                if "otg_iterations" in yml_dict:
                    self.otg_iterations = yml_dict["otg_iterations"]
                if "otg_grpc_transport" in yml_dict:
                    self.otg_grpc_transport = yml_dict["otg_grpc_transport"]
        else:
            log.info("Using default test configuration")

    def from_env(self):
        v = os.getenv("OTG_HOST")
        if v is not None:
            self.otg_host = v

        v = os.getenv("OTG_ITERATIONS")
        if v is not None:
            self.otg_iterations = int(v)

        v = os.getenv("OTG_GRPC_TRANSPORT")
        if v is not None:
            self.otg_grpc_transport = True if v in ["true", "True"] else False

        v = os.getenv("OTG_CAPTURE_CHECK")
        if v is not None:
            self.otg_capture_check = True if v in ["true", "True"] else False
