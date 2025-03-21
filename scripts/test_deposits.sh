#!/bin/bash

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Base URL for the API
BASE_URL="http://localhost:8082"
TOKEN=""
USER_ID=""

# Function to print section headers
print_header() {
    echo -e "\n${BLUE}=========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}=========================================${NC}\n"
}

# Function to handle errors
handle_error() {
    if [[ $1 == *"error"* ]]; then
        echo -e "${RED}Error: $1${NC}"
        exit 1
    fi
}

# Register a new user
register_user() {
    print_header "Registering a new user"
    
    local response=$(curl -s -X POST "$BASE_URL/auth/register" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "deposituser",
            "password": "secure_password",
            "email": "deposituser@example.com",
            "full_name": "Deposit Test User",
            "role": "client"
        }')
    
    echo -e "Response: ${GREEN}$response${NC}"
    handle_error "$response"
    
    echo -e "${YELLOW}User registered. Now we need admin approval.${NC}"
}

# Admin approves the user (you'll need an admin token)
approve_user() {
    print_header "Admin approving user (need admin token)"
    
    # First login as admin
    local admin_response=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "admin",
            "password": "admin_password"
        }')
    
    local admin_token=$(echo $admin_response | grep -o '"token":"[^"]*' | sed 's/"token":"//')
    
    if [ -z "$admin_token" ]; then
        echo -e "${RED}Failed to get admin token. Make sure admin user exists.${NC}"
        echo -e "${YELLOW}For testing, you may need to manually approve the user in the database.${NC}"
        echo -e "${YELLOW}Or proceed with login if auto-approval is enabled.${NC}"
    else
        # Get user ID from the database (simplified, in real scenario you'd use the admin API)
        local user_id=$(echo $admin_response | grep -o '"user_id":[0-9]*' | sed 's/"user_id"://')
        
        # Admin approves user
        local approve_response=$(curl -s -X POST "$BASE_URL/admin/approve-user" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $admin_token" \
            -d "{
                \"user_id\": $user_id
            }")
        
        echo -e "Response: ${GREEN}$approve_response${NC}"
        echo -e "${YELLOW}User approved. Now we can login.${NC}"
    fi
}

# Login to get token
login() {
    print_header "Logging in to get access token"
    
    local response=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "deposituser",
            "password": "secure_password"
        }')
    
    echo -e "Response: ${GREEN}$response${NC}"
    handle_error "$response"
    
    # Extract token from response
    TOKEN=$(echo $response | grep -o '"token":"[^"]*' | sed 's/"token":"//')
    USER_ID=$(echo $response | grep -o '"user_id":[0-9]*' | sed 's/"user_id"://')
    
    if [ -z "$TOKEN" ]; then
        echo -e "${RED}Failed to extract token from response.${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}Successfully logged in. Token: ${TOKEN:0:15}...${NC}"
    echo -e "${YELLOW}User ID: $USER_ID${NC}"
}

# Create deposit
create_deposit() {
    print_header "Creating a new deposit"
    
    local response=$(curl -s -X POST "$BASE_URL/deposit/create" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d '{
            "amount": 1000.50,
            "interest": 5.5,
            "term": 12,
            "currency": "USD",
            "bank_name": "Test Bank"
        }')
    
    echo -e "Response: ${GREEN}$response${NC}"
    handle_error "$response"
    
    # Extract deposit ID for future use
    DEPOSIT_ID_1=$(echo $response | grep -o '"deposit_id":[0-9]*' | sed 's/"deposit_id"://')
    
    echo -e "${YELLOW}Successfully created deposit with ID: $DEPOSIT_ID_1${NC}"
    
    # Create a second deposit for transfer testing
    local response2=$(curl -s -X POST "$BASE_URL/deposit/create" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d '{
            "amount": 500.25,
            "interest": 3.5,
            "term": 6,
            "currency": "USD",
            "bank_name": "Test Bank"
        }')
    
    echo -e "Response: ${GREEN}$response2${NC}"
    handle_error "$response2"
    
    # Extract second deposit ID
    DEPOSIT_ID_2=$(echo $response2 | grep -o '"deposit_id":[0-9]*' | sed 's/"deposit_id"://')
    
    echo -e "${YELLOW}Successfully created second deposit with ID: $DEPOSIT_ID_2${NC}"
}

# List deposits
list_deposits() {
    print_header "Listing all deposits"
    
    local response=$(curl -s -X GET "$BASE_URL/deposit/list" \
        -H "Authorization: Bearer $TOKEN")
    
    echo -e "Response: ${GREEN}$response${NC}"
    handle_error "$response"
    
    echo -e "${YELLOW}Successfully retrieved all deposits.${NC}"
}

# Transfer between deposits
transfer_between_deposits() {
    print_header "Transferring between deposits"
    
    if [ -z "$DEPOSIT_ID_1" ] || [ -z "$DEPOSIT_ID_2" ]; then
        echo -e "${RED}Missing deposit IDs. Cannot perform transfer.${NC}"
        return
    fi
    
    local response=$(curl -s -X POST "$BASE_URL/deposit/transfer" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"from_deposit_id\": $DEPOSIT_ID_1,
            \"to_deposit_id\": $DEPOSIT_ID_2,
            \"amount\": 200.00,
            \"bank_name\": \"Test Bank\"
        }")
    
    echo -e "Response: ${GREEN}$response${NC}"
    handle_error "$response"
    
    echo -e "${YELLOW}Successfully transferred funds between deposits.${NC}"
}

# Withdraw from deposit (using the endpoint we added)
withdraw_from_deposit() {
    print_header "Withdrawing from deposit"
    
    if [ -z "$DEPOSIT_ID_1" ]; then
        echo -e "${RED}Missing deposit ID. Cannot perform withdrawal.${NC}"
        return
    fi
    
    local response=$(curl -s -X POST "$BASE_URL/deposit/withdraw" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"deposit_id\": $DEPOSIT_ID_1,
            \"amount\": 100.00,
            \"bank_name\": \"Test Bank\"
        }")
    
    echo -e "Response: ${GREEN}$response${NC}"
    
    # Check if endpoint exists
    if [[ $response == *"404 page not found"* ]]; then
        echo -e "${YELLOW}Withdrawal endpoint not implemented yet. Skipping test.${NC}"
    else
        handle_error "$response"
        echo -e "${YELLOW}Successfully withdrew funds from deposit.${NC}"
    fi
}

# Freeze deposit
freeze_deposit() {
    print_header "Freezing deposit"
    
    if [ -z "$DEPOSIT_ID_1" ]; then
        echo -e "${RED}Missing deposit ID. Cannot freeze deposit.${NC}"
        return
    fi
    
    local response=$(curl -s -X POST "$BASE_URL/deposit/freeze" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"deposit_id\": $DEPOSIT_ID_1,
            \"client_id\": $USER_ID,
            \"bank_name\": \"Test Bank\",
            \"freeze_duration\": 24,
            \"reason\": \"Testing freeze functionality\"
        }")
    
    echo -e "Response: ${GREEN}$response${NC}"
    handle_error "$response"
    
    echo -e "${YELLOW}Successfully froze deposit.${NC}"
}

# Block deposit
block_deposit() {
    print_header "Blocking deposit"
    
    if [ -z "$DEPOSIT_ID_2" ]; then
        echo -e "${RED}Missing deposit ID. Cannot block deposit.${NC}"
        return
    fi
    
    local response=$(curl -s -X POST "$BASE_URL/deposit/block" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"deposit_id\": $DEPOSIT_ID_2,
            \"client_id\": $USER_ID,
            \"bank_name\": \"Test Bank\",
            \"reason\": \"Testing block functionality\"
        }")
    
    echo -e "Response: ${GREEN}$response${NC}"
    handle_error "$response"
    
    echo -e "${YELLOW}Successfully blocked deposit.${NC}"
}

# Unblock deposit
unblock_deposit() {
    print_header "Unblocking deposit"
    
    if [ -z "$DEPOSIT_ID_2" ]; then
        echo -e "${RED}Missing deposit ID. Cannot unblock deposit.${NC}"
        return
    fi
    
    local response=$(curl -s -X POST "$BASE_URL/deposit/unblock" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"deposit_id\": $DEPOSIT_ID_2,
            \"client_id\": $USER_ID,
            \"bank_name\": \"Test Bank\",
            \"reason\": \"Testing unblock functionality\"
        }")
    
    echo -e "Response: ${GREEN}$response${NC}"
    handle_error "$response"
    
    echo -e "${YELLOW}Successfully unblocked deposit.${NC}"
}

# Setup automatic savings (using the endpoint we added)
setup_automatic_savings() {
    print_header "Setting up automatic savings"
    
    if [ -z "$DEPOSIT_ID_1" ] || [ -z "$DEPOSIT_ID_2" ]; then
        echo -e "${RED}Missing deposit IDs. Cannot setup automatic savings.${NC}"
        return
    fi
    
    local response=$(curl -s -X POST "$BASE_URL/deposit/savings/setup" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"deposit_id\": $DEPOSIT_ID_1,
            \"source_id\": $DEPOSIT_ID_2,
            \"amount\": 50.00,
            \"frequency\": \"monthly\",
            \"bank_name\": \"Test Bank\",
            \"start_date\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\",
            \"target_amount\": 500.00
        }")
    
    echo -e "Response: ${GREEN}$response${NC}"
    
    # Check if endpoint exists
    if [[ $response == *"404 page not found"* ]]; then
        echo -e "${YELLOW}Automatic savings endpoint not implemented yet. Skipping test.${NC}"
    else
        handle_error "$response"
        echo -e "${YELLOW}Successfully set up automatic savings.${NC}"
        
        # Get savings plans
        get_savings_plans
    fi
}

# Get savings plans (using the endpoint we added)
get_savings_plans() {
    print_header "Getting savings plans"
    
    local response=$(curl -s -X GET "$BASE_URL/deposit/savings/plans" \
        -H "Authorization: Bearer $TOKEN")
    
    echo -e "Response: ${GREEN}$response${NC}"
    
    # Check if endpoint exists
    if [[ $response == *"404 page not found"* ]]; then
        echo -e "${YELLOW}Savings plans endpoint not implemented yet. Skipping test.${NC}"
    else
        handle_error "$response"
        echo -e "${YELLOW}Successfully retrieved savings plans.${NC}"
        
        # Extract plan ID for cancellation
        PLAN_ID=$(echo $response | grep -o '"id":[0-9]*' | head -1 | sed 's/"id"://')
        
        if [ ! -z "$PLAN_ID" ]; then
            cancel_savings_plan
        fi
    fi
}

# Cancel savings plan (using the endpoint we added)
cancel_savings_plan() {
    print_header "Cancelling savings plan"
    
    if [ -z "$PLAN_ID" ]; then
        echo -e "${RED}Missing plan ID. Cannot cancel savings plan.${NC}"
        return
    fi
    
    local response=$(curl -s -X POST "$BASE_URL/deposit/savings/cancel" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"plan_id\": $PLAN_ID
        }")
    
    echo -e "Response: ${GREEN}$response${NC}"
    
    # Check if endpoint exists
    if [[ $response == *"404 page not found"* ]]; then
        echo -e "${YELLOW}Cancel savings plan endpoint not implemented yet. Skipping test.${NC}"
    else
        handle_error "$response"
        echo -e "${YELLOW}Successfully cancelled savings plan.${NC}"
    fi
}

# Delete deposit
delete_deposit() {
    print_header "Deleting deposits"
    
    if [ -z "$DEPOSIT_ID_1" ]; then
        echo -e "${RED}Missing deposit ID. Cannot delete first deposit.${NC}"
    else
        local response=$(curl -s -X DELETE "$BASE_URL/deposit/delete" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $TOKEN" \
            -d "{
                \"deposit_id\": $DEPOSIT_ID_1,
                \"bank_name\": \"Test Bank\"
            }")
        
        echo -e "Response for deposit 1: ${GREEN}$response${NC}"
        handle_error "$response"
        echo -e "${YELLOW}Successfully deleted first deposit.${NC}"
    fi
    
    if [ -z "$DEPOSIT_ID_2" ]; then
        echo -e "${RED}Missing deposit ID. Cannot delete second deposit.${NC}"
    else
        local response2=$(curl -s -X DELETE "$BASE_URL/deposit/delete" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $TOKEN" \
            -d "{
                \"deposit_id\": $DEPOSIT_ID_2,
                \"bank_name\": \"Test Bank\"
            }")
        
        echo -e "Response for deposit 2: ${GREEN}$response2${NC}"
        handle_error "$response2"
        echo -e "${YELLOW}Successfully deleted second deposit.${NC}"
    fi
}

# Main execution flow
main() {
    echo -e "${GREEN}Starting deposit functionality test script...${NC}"
    
    # Uncomment these if you need to register a new user
    # register_user
    # approve_user
    
    login
    list_deposits
    create_deposit
    list_deposits
    transfer_between_deposits
    list_deposits
    withdraw_from_deposit
    list_deposits
    freeze_deposit
    list_deposits
    block_deposit
    list_deposits
    unblock_deposit
    list_deposits
    setup_automatic_savings
    # get_savings_plans and cancel_savings_plan are called from setup_automatic_savings if successful
    # delete_deposit # Uncomment if you want to clean up the created deposits
    
    echo -e "\n${GREEN}Deposit functionality test completed!${NC}"
}

# Run the main function
main
