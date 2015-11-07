FROM centurylink/ca-certs
MAINTAINER Anders Pearson <anders@columbia.edu>
COPY frontdesk /
EXPOSE 8890
CMD ["/frontdesk"]
