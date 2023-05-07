# Scanning

## Repository Advanced Scans

{% swagger method="post" path="" baseUrl="/api/v1/repository/advancedScan/scan" summary="Invokes JAS Exposures and Contextual Analysis scanning of a repository." %}
{% swagger-description %}

{% endswagger-description %}

{% swagger-parameter in="body" name="repository" type="string" required="true" %}
The name of the repository to scan.
{% endswagger-parameter %}

{% swagger-response status="200: OK" description="" %}

{% endswagger-response %}
{% endswagger %}
