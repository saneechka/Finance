# Deposit Testing Scripts

This directory contains scripts to test the deposit functionality of the banking system.

## Bash Script (`test_deposits.sh`)

A comprehensive bash script that tests all deposit-related operations:

1. User registration/login
2. Creating deposits
3. Listing deposits
4. Transferring between deposits
5. Withdrawing from deposits
6. Freezing/unfreezing deposits
7. Blocking/unblocking deposits
8. Setting up automatic savings
9. Managing savings plans
10. Deleting deposits

### Usage

1. Make the script executable:
   ```bash
   chmod +x test_deposits.sh
   ```

2. Run the script:
   ```bash
   ./test_deposits.sh
   ```

3. Customize as needed:
   - Edit user credentials
   - Modify deposit amounts and parameters
   - Comment out operations you don't want to test

## HTTP Requests (`deposit_requests.http`)

A collection of HTTP requests for testing with tools like VS Code's REST Client extension or Postman.

### Usage with VS Code REST Client

1. Install the "REST Client" extension in VS Code
2. Open the `deposit_requests.http` file
3. After logging in, copy the token from the response
4. Update the `@token` variable at the top of the file
5. Click "Send Request" above each request to execute it

### Usage with Postman

1. Import the requests into Postman
2. Create a collection variable named `token` and set it after login
3. Execute the requests in order

## Notes

- Some endpoints may not be implemented yet, the script handles these cases gracefully
- Ensure the server is running at `http://localhost:8082` before running the scripts
- You may need admin privileges to approve new users
