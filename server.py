# noinspection PyPackageRequirements
from flask import Flask, render_template, request, jsonify

from database import init_db, add_email

app = Flask(__name__)
init_db()

@app.route("/")
def index():
    return render_template("index.html")

@app.route("/subscribe", methods=["POST"])
def subscribe():
    data = request.get_json()
    email = data.get('email')
    if not email:
        return jsonify({"error": "No email provided"}), 400

    if add_email(email):
        return jsonify({"message": "Email has been added"}), 201
    else:
        return jsonify({"error": "Email already exists"}), 400

if __name__ == "__main__":
    app.run(debug=True)
