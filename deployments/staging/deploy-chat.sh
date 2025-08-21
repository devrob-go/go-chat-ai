#!/bin/bash

# Chat Service Deployment Script
# This script deploys the chat service to Kubernetes using Helm

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
CHART_NAME="chat-service"
NAMESPACE="${NAMESPACE:-default}"
VALUES_FILE="charts/chat-service/values.yaml"
DRY_RUN="${DRY_RUN:-false}"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if helm is installed
    if ! command -v helm &> /dev/null; then
        print_error "helm is not installed. Please install helm first."
        exit 1
    fi
    
    # Check if we can connect to the cluster
    if ! kubectl cluster-info &> /dev/null; then
        print_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi
    
    print_status "Prerequisites check passed."
}

# Function to validate configuration
validate_config() {
    print_status "Validating configuration..."
    
    # Check if values file exists
    if [[ ! -f "$VALUES_FILE" ]]; then
        print_error "Values file $VALUES_FILE not found."
        exit 1
    fi
    
    # Check required environment variables
    if [[ -z "$OPENAI_API_KEY" ]]; then
        print_warning "OPENAI_API_KEY environment variable is not set."
        print_warning "You may need to set it in the values.yaml file or as an environment variable."
    fi
    
    print_status "Configuration validation completed."
}

# Function to deploy the service
deploy_service() {
    print_status "Deploying chat service to namespace: $NAMESPACE"
    
    # Create namespace if it doesn't exist
    if [[ "$NAMESPACE" != "default" ]]; then
        kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    fi
    
    # Deploy using Helm
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "Running in dry-run mode..."
        helm upgrade --install "$CHART_NAME" ./charts/chat-service \
            --namespace "$NAMESPACE" \
            --values "$VALUES_FILE" \
            --dry-run
    else
        helm upgrade --install "$CHART_NAME" ./charts/chat-service \
            --namespace "$NAMESPACE" \
            --values "$VALUES_FILE" \
            --wait \
            --timeout 5m
    fi
    
    print_status "Deployment completed successfully!"
}

# Function to check deployment status
check_status() {
    print_status "Checking deployment status..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "Skipping status check in dry-run mode."
        return
    fi
    
    # Wait for pods to be ready
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=chat-service \
        --namespace "$NAMESPACE" \
        --timeout=300s
    
    # Show pod status
    kubectl get pods --namespace "$NAMESPACE" -l app.kubernetes.io/name=chat-service
    
    # Show service status
    kubectl get services --namespace "$NAMESPACE" -l app.kubernetes.io/name=chat-service
    
    print_status "Status check completed."
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -n, --namespace NAMESPACE  Kubernetes namespace (default: default)"
    echo "  -f, --values FILE          Values file (default: charts/chat-service/values.yaml)"
    echo "  -d, --dry-run              Run in dry-run mode"
    echo "  -h, --help                 Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  OPENAI_API_KEY             OpenAI API key for the service"
    echo "  NAMESPACE                  Kubernetes namespace"
    echo "  DRY_RUN                    Set to 'true' for dry-run mode"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Deploy to default namespace"
    echo "  $0 -n chat-system                     # Deploy to chat-system namespace"
    echo "  $0 -d                                 # Dry-run deployment"
    echo "  OPENAI_API_KEY=sk-... $0             # Deploy with API key"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -f|--values)
            VALUES_FILE="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN="true"
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    print_status "Starting chat service deployment..."
    
    check_prerequisites
    validate_config
    deploy_service
    check_status
    
    print_status "Deployment process completed successfully!"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        echo ""
        echo "Next steps:"
        echo "1. Verify the service is running: kubectl get pods -n $NAMESPACE"
        echo "2. Check service logs: kubectl logs -n $NAMESPACE -l app.kubernetes.io/name=chat-service"
        echo "3. Test the service endpoints"
        echo "4. Monitor metrics and logs"
    fi
}

# Run main function
main "$@"
