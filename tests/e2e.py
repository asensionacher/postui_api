import requests
import json
import pytest

# Define the base URL
BASE_URL = "http://localhost:8001/api/v1"

headers = {"X-API-Key": API_KEY, "Content-Type": "application/json"}

# Register a new user and log in to get a JWT token
def get_jwt_token():
    # User registration (assuming it doesn't need authentication)
    register_url = f"{BASE_URL}/register"
    register_data = json.dumps({"username": "user1", "password": "securepassword"})
    register_response = requests.post(register_url, headers=headers, data=register_data)
    
    assert register_response.status_code in [200, 201], f"User registration failed: {register_response.status_code}"
    
    # User login
    login_url = f"{BASE_URL}/login"
    login_data = json.dumps({"username": "user1", "password": "securepassword"})
    login_response = requests.post(login_url, headers=headers, data=login_data)
    
    assert login_response.status_code == 200, f"User login failed: {login_response.status_code}"
    
    # Extracting JWT token from login response
    jwt_token = login_response.json().get('token')  
    assert jwt_token is not None, "JWT token not found in login response"
    
    return jwt_token

# Create a new product
def create_product(jwt_token):
    url = f"{BASE_URL}/products"
    auth_headers = {**headers, "Authorization": f"Bearer {jwt_token}"}
    data = json.dumps({"author": "Jane Doe", "title": "New Product Title"})
    response = requests.post(url, headers=auth_headers, data=data)
    
    assert response.status_code == 201, f"Failed to create product: {response.status_code}"
    return response.json()

# Get all products
def get_products(jwt_token):
    url = f"{BASE_URL}/products"
    auth_headers = {**headers, "Authorization": f"Bearer {jwt_token}"}
    response = requests.get(url, headers=auth_headers)
    
    assert response.status_code == 200, f"Failed to get products: {response.status_code}"
    return response.json()

# Get a specific product by ID
def get_product(product_id, jwt_token):
    url = f"{BASE_URL}/products/{product_id}"
    auth_headers = {**headers, "Authorization": f"Bearer {jwt_token}"}
    response = requests.get(url, headers=auth_headers)
    
    assert response.status_code == 200, f"Failed to get product: {response.status_code}"
    return response.json()

# Update a product
def update_product(product_id, jwt_token):
    url = f"{BASE_URL}/products/{product_id}"
    auth_headers = {**headers, "Authorization": f"Bearer {jwt_token}"}
    data = json.dumps({"author": "John Smith", "title": "Updated Product Title"})
    response = requests.put(url, headers=auth_headers, data=data)
    
    assert response.status_code == 200, f"Failed to update product: {response.status_code}"
    return response.json()

# Delete a product
def delete_product(product_id, jwt_token):
    url = f"{BASE_URL}/products/{product_id}"
    auth_headers = {**headers, "Authorization": f"Bearer {jwt_token}"}
    response = requests.delete(url, headers=auth_headers)
    
    assert response.status_code == 204, f"Failed to delete product: {response.status_code}"

# Tests using pytest
@pytest.fixture(scope="module")
def jwt_token():
    return get_jwt_token()

def test_create_product(jwt_token):
    product = create_product(jwt_token)
    assert product['data']['author'] == "Jane Doe"
    assert product['data']['title'] == "New Product Title"
    return product

def test_get_products(jwt_token):
    products = get_products(jwt_token)
    assert isinstance(products, dict), "Products response is not a dictionary"
    assert 'data' in products, "Products data not found in response"

def test_get_product(jwt_token):
    product = create_product(jwt_token)
    product_id = product['data']['id']
    fetched_product = get_product(product_id, jwt_token)
    assert fetched_product['data']['id'] == product_id, "Fetched product ID does not match created product ID"

def test_update_product(jwt_token):
    product = create_product(jwt_token)
    product_id = product['data']['id']
    updated_product = update_product(product_id, jwt_token)
    assert updated_product['data']['author'] == "John Smith", "Product author not updated"
    assert updated_product['data']['title'] == "Updated Product Title", "Product title not updated"

def test_delete_product(jwt_token):
    product = create_product(jwt_token)
    product_id = product['data']['id']
    delete_product(product_id, jwt_token)
    with pytest.raises(AssertionError):
        get_product(product_id, jwt_token)  # This should raise an error since the product is deleted
