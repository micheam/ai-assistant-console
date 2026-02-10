# Integration Test Dockerfile for AICO Installation
# Tests the install.sh script in a clean Ubuntu environment

FROM ubuntu:22.04

# Install required dependencies
RUN apt-get update && \
    apt-get install -y \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create a non-root user for testing
RUN useradd -m -s /bin/bash testuser

# Switch to test user
USER testuser
WORKDIR /home/testuser

# Copy the installation script
COPY --chown=testuser:testuser install.sh /home/testuser/install.sh

# Set up PATH for the test
ENV PATH="/home/testuser/.local/bin:${PATH}"

# Default command runs the test
CMD ["/bin/bash"]
