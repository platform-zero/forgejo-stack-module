package org.webservices.testrunner.suites

import io.ktor.client.statement.*
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

    test("Forgejo version API reports the running server version") {
        val response = client.getRawResponse("${env.endpoints.forgejo}/api/v1/version")
        if (response.status == io.ktor.http.HttpStatusCode.OK) {
            response.bodyAsText() shouldContain "version"
        } else {
            requireAuthBoundary(response, "Forgejo version API")
        }
    }
}
