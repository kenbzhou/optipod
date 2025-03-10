FROM registry.k8s.io/kube-scheduler:v1.28.3

# Copy the plugin
COPY customscorer.so /usr/local/bin/plugins/

# Copy the startup script and make it executable in one step
COPY --chmod=755 start-scheduler.sh /usr/local/bin/

# Set the entrypoint to our script
ENTRYPOINT ["/usr/local/bin/start-scheduler.sh"]