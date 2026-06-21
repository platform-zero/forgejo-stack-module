package org.webservices.testrunner.suites

import org.webservices.testrunner.framework.*

suspend fun TestRunner.forgejoProductivityTests() = suite("Forgejo Productivity Tests") {
test("Forgejo git server web interface is healthy") {
        val response = client.getRawResponse("${env.endpoints.forgejo}/")
        requireOkOrRedirectResponse(response, "Forgejo web interface")
    }

    test("Forgejo web interface loads") {
        val response = client.getRawResponse("${env.endpoints.forgejo}/")
        requireOkOrRedirectResponse(response, "Forgejo web interface")
    }

    test("Forgejo API enforces authentication") {
        val response = client.getRawResponse("${env.endpoints.forgejo}/api/v1/version")
        requireAuthBoundary(response, "Forgejo version API")
    }
}
