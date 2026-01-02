#!/bin/bash
# Helper script to get the GitHub Project v2 node ID
# This requires the 'project' scope on your GitHub token

# Usage: ./get-project-id.sh <org> <project-number>
# Example: ./get-project-id.sh mondoohq 14

if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Usage: $0 <org> <project-number>"
    echo "Example: $0 mondoohq 14"
    exit 1
fi

ORG=$1
PROJECT_NUM=$2

# Note: This requires the 'project' scope on your token
# Run: gh auth refresh -s project

GH_TOKEN="${GITHUB_TOKEN_ASSISTANT_MONDOOHQ}" gh api graphql -f query="
{
  organization(login: \"${ORG}\") {
    projectV2(number: ${PROJECT_NUM}) {
      id
      title
      number
      url
    }
  }
}
" | jq -r '.data.organization.projectV2 | "Project ID: \(.id)\nTitle: \(.title)\nURL: \(.url)"'
