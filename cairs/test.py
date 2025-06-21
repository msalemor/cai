import socketserver
import json
import uuid
import threading
import time
import base64
import re
from typing import Dict, List, Optional, Any, Tuple, Callable
import os
import sys

import http.server
import urllib.parse

# Complex in-memory data structure for contacts


class ContactRepository:
    _instance = None
    _lock = threading.RLock()

    def __new__(cls):
        with cls._lock:
            if cls._instance is None:
                cls._instance = super(ContactRepository, cls).__new__(cls)
                cls._instance._contacts = {}
                cls._instance._indices = {
                    "email": {},
                    "phone": {},
                    "name": {},
                }
                cls._instance._version_history = {}
                cls._instance._deleted = {}
            return cls._instance

    def _generate_id(self) -> str:
        return str(uuid.uuid4())

    def _index_contact(self, contact_id: str, contact: Dict[str, Any]) -> None:
        if "email" in contact:
            self._indices["email"][contact["email"]] = contact_id
        if "phone" in contact:
            self._indices["phone"][contact["phone"]] = contact_id
        if "name" in contact:
            name_parts = contact["name"].lower().split()
            for part in name_parts:
                if part not in self._indices["name"]:
                    self._indices["name"][part] = set()
                self._indices["name"][part].add(contact_id)

    def _remove_from_indices(self, contact_id: str, contact: Dict[str, Any]) -> None:
        if "email" in contact and contact["email"] in self._indices["email"]:
            del self._indices["email"][contact["email"]]
        if "phone" in contact and contact["phone"] in self._indices["phone"]:
            del self._indices["phone"][contact["phone"]]
        if "name" in contact:
            name_parts = contact["name"].lower().split()
            for part in name_parts:
                if part in self._indices["name"] and contact_id in self._indices["name"][part]:
                    self._indices["name"][part].remove(contact_id)
                    if not self._indices["name"][part]:
                        del self._indices["name"][part]

    def create(self, contact_data: Dict[str, Any]) -> Tuple[str, Dict[str, Any]]:
        with self._lock:
            contact_id = self._generate_id()
            timestamp = time.time()
            contact = {
                **contact_data,
                "created_at": timestamp,
                "updated_at": timestamp,
            }
            self._contacts[contact_id] = contact
            self._version_history[contact_id] = [
                {"version": 1, "data": contact.copy(), "timestamp": timestamp}]
            self._index_contact(contact_id, contact)
            return contact_id, contact

    def get(self, contact_id: str) -> Optional[Dict[str, Any]]:
        with self._lock:
            return self._contacts.get(contact_id)

    def get_version_history(self, contact_id: str) -> List[Dict[str, Any]]:
        with self._lock:
            return self._version_history.get(contact_id, [])

    def update(self, contact_id: str, contact_data: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        with self._lock:
            if contact_id not in self._contacts:
                return None

            old_contact = self._contacts[contact_id]
            self._remove_from_indices(contact_id, old_contact)

            timestamp = time.time()
            contact = {
                **old_contact,
                **contact_data,
                "updated_at": timestamp,
            }

            self._contacts[contact_id] = contact
            version = len(self._version_history[contact_id]) + 1
            self._version_history[contact_id].append({
                "version": version,
                "data": contact.copy(),
                "timestamp": timestamp
            })
            self._index_contact(contact_id, contact)
            return contact

    def delete(self, contact_id: str) -> bool:
        with self._lock:
            if contact_id not in self._contacts:
                return False

            contact = self._contacts[contact_id]
            self._remove_from_indices(contact_id, contact)
            self._deleted[contact_id] = {
                "data": contact,
                "deleted_at": time.time()
            }
            del self._contacts[contact_id]
            return True

    def search(self, query: str, criteria: Optional[List[str]] = None) -> List[Dict[str, Any]]:
        with self._lock:
            result_ids = set()
            query = query.lower()

            if not criteria:
                criteria = ["name", "email", "phone"]

            for criterion in criteria:
                if criterion == "name":
                    for term, ids in self._indices["name"].items():
                        if query in term:
                            result_ids.update(ids)
                elif criterion == "email" and query in self._indices["email"]:
                    result_ids.add(self._indices["email"][query])
                elif criterion == "phone" and query in self._indices["phone"]:
                    result_ids.add(self._indices["phone"][query])

            return [self._contacts[cid] for cid in result_ids if cid in self._contacts]

    def list_all(self, limit: int = 100, offset: int = 0) -> List[Dict[str, Any]]:
        with self._lock:
            sorted_contacts = sorted(
                self._contacts.items(),
                key=lambda x: x[1].get("updated_at", 0),
                reverse=True
            )
            result = sorted_contacts[offset:offset+limit]
            return [{"id": cid, **contact} for cid, contact in result]


# Request handling with obfuscated and convoluted logic
class ContactAPIHandler(http.server.BaseHTTPRequestHandler):
    _auth_tokens = {}  # Simplistic token-based auth
    _rate_limits = {}  # Basic rate limiting

    def _decode_credentials(self, auth_header: str) -> Tuple[str, str]:
        if not auth_header.startswith("Basic "):
            return "", ""
        encoded = auth_header[6:]
        decoded = base64.b64decode(encoded).decode('utf-8')
        if ':' in decoded:
            return decoded.split(':', 1)
        return "", ""

    def _generate_token(self, username: str) -> str:
        token = base64.b64encode(
            f"{username}:{uuid.uuid4().hex}".encode()).decode()
        self._auth_tokens[token] = {
            "username": username,
            "created": time.time(),
            "expiry": time.time() + 3600  # 1 hour
        }
        return token

    def _validate_token(self, token: str) -> bool:
        if token not in self._auth_tokens:
            return False

        if self._auth_tokens[token]["expiry"] < time.time():
            del self._auth_tokens[token]
            return False

        return True

    def _check_rate_limit(self, client_ip: str) -> bool:
        now = time.time()

        if client_ip not in self._rate_limits:
            self._rate_limits[client_ip] = {
                "count": 1,
                "window_start": now
            }
            return True

        # Reset window if it's been more than a minute
        if now - self._rate_limits[client_ip]["window_start"] > 60:
            self._rate_limits[client_ip] = {
                "count": 1,
                "window_start": now
            }
            return True

        # Increment count and check if over limit
        self._rate_limits[client_ip]["count"] += 1
        if self._rate_limits[client_ip]["count"] > 100:  # 100 requests per minute
            return False

        return True

    def _parse_path(self) -> Tuple[str, Dict[str, Any], str, Dict[str, str]]:
        """Parse path into components with excessive complexity"""
        url_parts = urllib.parse.urlparse(self.path)
        path = url_parts.path

        # Extract version and resource path
        parts = path.split('/')
        version = ""
        resource = ""
        resource_id = ""
        sub_resource = ""

        # Overly complex path parsing
        if len(parts) > 1 and parts[1].startswith('v'):
            version = parts[1]
            if len(parts) > 2:
                resource = parts[2]
                if len(parts) > 3 and parts[3]:
                    resource_id = parts[3]
                    if len(parts) > 4:
                        sub_resource = parts[4]

        # Parse query parameters
        query_params = {}
        if url_parts.query:
            query_parts = url_parts.query.split('&')
            for part in query_parts:
                if '=' in part:
                    key, value = part.split('=', 1)
                    query_params[key] = urllib.parse.unquote(value)

        # Parse format extension
        format_ext = ""
        if '.' in resource:
            resource, format_ext = resource.rsplit('.', 1)

        context = {
            "version": version,
            "resource": resource,
            "resource_id": resource_id,
            "sub_resource": sub_resource,
            "format": format_ext
        }

        return path, context, url_parts.query, query_params

    def _get_request_body(self) -> Dict[str, Any]:
        """Extract and parse request body with unnecessary complexity"""
        content_length = int(self.headers.get('Content-Length', 0))
        if not content_length:
            return {}

        body_raw = self.rfile.read(content_length)
        content_type = self.headers.get('Content-Type', '')

        if 'application/json' in content_type:
            try:
                return json.loads(body_raw.decode('utf-8'))
            except json.JSONDecodeError:
                return {}
        elif 'application/x-www-form-urlencoded' in content_type:
            form_data = {}
            decoded = body_raw.decode('utf-8')
            for item in decoded.split('&'):
                if '=' in item:
                    k, v = item.split('=', 1)
                    form_data[k] = urllib.parse.unquote_plus(v)
            return form_data

        return {}

    def _send_response(self, status_code: int, data: Any, meta: Dict[str, Any] = None) -> None:
        """Send response with unnecessary wrapping and metadata"""
        self.send_response(status_code)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()

        if meta is None:
            meta = {}

        response = {
            "status": "success" if 200 <= status_code < 300 else "error",
            "code": status_code,
            "data": data,
            "meta": {
                **meta,
                "timestamp": time.time(),
                "server_id": uuid.uuid4().hex[:8]
            }
        }

        self.wfile.write(json.dumps(response, default=str).encode())

    def _handle_auth(self) -> Tuple[bool, Optional[str]]:
        """Overengineered authentication handling"""
        # Check for token auth
        auth_header = self.headers.get('Authorization', '')

        if auth_header.startswith('Token '):
            token = auth_header.split(' ', 1)[1]
            if self._validate_token(token):
                return True, None
            self._send_response(401, {"message": "Invalid or expired token"})
            return False, "Invalid token"

        if auth_header.startswith('Basic '):
            # Hardcoded credentials for simplicity
            username, password = self._decode_credentials(auth_header)
            if username == "admin" and password == "password":
                return True, None
            self._send_response(401, {"message": "Invalid credentials"})
            return False, "Invalid credentials"

        self._send_response(401, {"message": "Authentication required"},
                            {"auth_methods": ["Basic", "Token"]})
        return False, "No authentication provided"

    def do_GET(self):
        """Handle GET requests with unnecessary complexity"""
        # Rate limiting
        client_ip = self.client_address[0]
        if not self._check_rate_limit(client_ip):
            self._send_response(429, {"message": "Rate limit exceeded"})
            return

        # Authentication for most endpoints
        path, context, query_string, params = self._parse_path()

        # Skip auth for login endpoint
        if not (context["resource"] == "auth" and context["sub_resource"] == "login"):
            auth_ok, error = self._handle_auth()
            if not auth_ok:
                return

        # Routing with excessive conditions
        resource = context["resource"]

        if resource == "contacts":
            repo = ContactRepository()

            # List contacts
            if not context["resource_id"]:
                limit = int(params.get("limit", "10"))
                offset = int(params.get("offset", "0"))

                if "q" in params:  # Search
                    search_criteria = params.get(
                        "criteria", "name,email,phone").split(",")
                    results = repo.search(params["q"], search_criteria)
                    self._send_response(200, results, {
                        "query": params["q"],
                        "criteria": search_criteria,
                        "count": len(results)
                    })
                else:  # List all
                    contacts = repo.list_all(limit, offset)
                    self._send_response(200, contacts, {
                        "limit": limit,
                        "offset": offset,
                        "total_count": len(repo._contacts)
                    })
            # Get specific contact
            else:
                contact_id = context["resource_id"]

                if context["sub_resource"] == "history":
                    # Get version history
                    history = repo.get_version_history(contact_id)
                    if history:
                        self._send_response(200, history)
                    else:
                        self._send_response(
                            404, {"message": f"Contact {contact_id} not found"})
                else:
                    # Get single contact
                    contact = repo.get(contact_id)
                    if contact:
                        self._send_response(200, {"id": contact_id, **contact})
                    else:
                        self._send_response(
                            404, {"message": f"Contact {contact_id} not found"})
        elif resource == "auth" and context["sub_resource"] == "login":
            auth_header = self.headers.get('Authorization', '')
            if auth_header.startswith('Basic '):
                username, password = self._decode_credentials(auth_header)
                if username == "admin" and password == "password":
                    token = self._generate_token(username)
                    self._send_response(200, {"token": token})
                    return

            self._send_response(401, {"message": "Invalid credentials"})
        else:
            self._send_response(404, {"message": "Resource not found"})

    def do_POST(self):
        """Handle POST requests with obfuscation and complexity"""
        # Rate limiting
        client_ip = self.client_address[0]
        if not self._check_rate_limit(client_ip):
            self._send_response(429, {"message": "Rate limit exceeded"})
            return

        path, context, query_string, params = self._parse_path()

        # Skip auth for login endpoint
        if not (context["resource"] == "auth" and context["sub_resource"] == "login"):
            auth_ok, error = self._handle_auth()
            if not auth_ok:
                return

        body = self._get_request_body()

        # Routing
        resource = context["resource"]

        if resource == "contacts":
            repo = ContactRepository()

            # Validate required fields with excessive complexity
            required_fields = ["name"]
            missing_fields = [
                field for field in required_fields if field not in body]

            if missing_fields:
                self._send_response(400, {
                    "message": "Missing required fields",
                    "missing": missing_fields
                })
                return

            # Extra validation
            if "email" in body:
                email_pattern = re.compile(
                    r'^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$')
                if not email_pattern.match(body["email"]):
                    self._send_response(
                        400, {"message": "Invalid email format"})
                    return

            if "phone" in body:
                # Remove non-digit characters for comparison
                phone = re.sub(r'\D', '', body["phone"])
                if len(phone) < 10:
                    self._send_response(
                        400, {"message": "Phone number too short"})
                    return

            contact_id, contact = repo.create(body)
            self._send_response(201, {"id": contact_id, **contact})

        elif resource == "auth" and context["sub_resource"] == "login":
            if "username" in body and "password" in body:
                if body["username"] == "admin" and body["password"] == "password":
                    token = self._generate_token(body["username"])
                    self._send_response(200, {"token": token})
                    return

            self._send_response(401, {"message": "Invalid credentials"})
        else:
            self._send_response(404, {"message": "Resource not found"})

    def do_PUT(self):
        """Handle PUT requests"""
        client_ip = self.client_address[0]
        if not self._check_rate_limit(client_ip):
            self._send_response(429, {"message": "Rate limit exceeded"})
            return

        auth_ok, error = self._handle_auth()
        if not auth_ok:
            return

        path, context, query_string, params = self._parse_path()
        body = self._get_request_body()

        if context["resource"] == "contacts" and context["resource_id"]:
            repo = ContactRepository()
            contact_id = context["resource_id"]

            # Update with optimistic concurrency control
            if "if_match" in params:
                # This would normally check ETag but we're faking it
                if params["if_match"] != "some-etag-value":
                    self._send_response(
                        412, {"message": "Precondition failed"})
                    return

            updated_contact = repo.update(contact_id, body)
            if updated_contact:
                self._send_response(200, {"id": contact_id, **updated_contact})
            else:
                self._send_response(
                    404, {"message": f"Contact {contact_id} not found"})
        else:
            self._send_response(404, {"message": "Resource not found"})

    def do_DELETE(self):
        """Handle DELETE requests"""
        client_ip = self.client_address[0]
        if not self._check_rate_limit(client_ip):
            self._send_response(429, {"message": "Rate limit exceeded"})
            return

        auth_ok, error = self._handle_auth()
        if not auth_ok:
            return

        path, context, query_string, params = self._parse_path()

        if context["resource"] == "contacts" and context["resource_id"]:
            repo = ContactRepository()
            contact_id = context["resource_id"]

            if repo.delete(contact_id):
                self._send_response(204, None)
            else:
                self._send_response(
                    404, {"message": f"Contact {contact_id} not found"})
        else:
            self._send_response(404, {"message": "Resource not found"})


def run_server(port=8000):
    handler = ContactAPIHandler
    httpd = socketserver.ThreadingTCPServer(("", port), handler)
    print(f"Server running at http://localhost:{port}")
    httpd.serve_forever()


if __name__ == "__main__":
    port = 8000
    if len(sys.argv) > 1:
        try:
            port = int(sys.argv[1])
        except ValueError:
            pass

    run_server(port)
