import snappi
import pytest


@pytest.fixture(scope="session")
def otg():
    return snappi.api(location="https://localhost", verify=False)
