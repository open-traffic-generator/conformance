import pexpect
import yaml

# Load versions.yaml file
with open("versions.yaml", "r") as yaml_file:
    versions_data = yaml.safe_load(yaml_file)
    uhd400_version = next((image.get("tag") for image in versions_data.get("images", []) if image.get("name") == "uhd400"), "")
    uhd400_path = next((image.get("path") for image in versions_data.get("images", []) if image.get("name") == "uhd400"), "")

# SSH into the remote server
ip_address="10.36.87.166"
password = "admin"
ssh_command = f"ssh admin@{ip_address}"
ssh_process = pexpect.spawn(ssh_command, timeout=None)

try:
    # Expect the RSA key fingerprint message or the password prompt
    index = ssh_process.expect(["RSA key fingerprint is .*", "password:", "yes/no/[fingerprint]"])

    if index == 0:
        print("Accepting RSA key fingerprint.")
        # Send 'yes' to confirm adding the host key
        ssh_process.sendline("yes")
        # Expect the password prompt
        ssh_process.expect("password:")
    elif index == 2:
        print("Confirming connection.")
        # Send 'yes' to confirm continuing connection
        ssh_process.sendline("yes")
        # Expect the password prompt
        ssh_process.expect("password:")

    # Send SSH password
    ssh_process.sendline(password)
    print("Logging in...")

    # Wait for the KCOS shell prompt
    ssh_process.expect("kcos-framework-shell-.*$")
    print("Logged in successfully.")

    try:
        # Execute curl command with uhd400_version
        ssh_process.sendline(f"curl -LOf https://{uhd400_path}/{uhd400_version}/artifacts.tar")
        index = ssh_process.expect([r"kcos-framework-shell-.*$", "404 Not Found", r"curl: \(.*"])  # Wait for the prompt or specific error messages
        if index == 0:
            print("Downloaded artifacts.")
            print("Command Output:")
            print(ssh_process.before.decode())  # Print command output
        elif index == 1:
            raise Exception("404 Not Found: The requested artifacts.tar file was not found.")
        elif index == 2:
            raise Exception("curl command encountered an error.")
    except Exception as e:
        print(f"Error during curl command: {e}")

    try:
        # Execute kcos deployment command
        ssh_process.sendline("kcos deployment offline-install artifacts.tar")
        index = ssh_process.expect(["Done", "ERROR"])  # Wait for the prompt or specific error messages
        if index == 0:
            print("Installed artifacts.")
            print("Command Output:")
            print(ssh_process.before.decode())  # Print command output
        elif index == 1:
            raise Exception("kcos deployment command encountered an error.")
    except Exception as e:
        print(f"Error during kcos deployment command: {e}")

    # Close the SSH connection
    ssh_process.sendline("exit")
    ssh_process.expect(pexpect.EOF)  # Wait for the SSH session to close completely

    print("SSH session closed successfully.")
except pexpect.exceptions.TIMEOUT as e:
    print(f"Timeout exceeded: {e}")
    # Close the SSH process if it's still open
    ssh_process.close()
except Exception as e:
    print(f"An error occurred: {e}")
    # Close the SSH process if it's still open
    ssh_process.close()
