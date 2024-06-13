from flask import Flask, request, jsonify
import spacy

# Load the spaCy model
nlp = spacy.load("en_core_web_sm")

app = Flask(__name__)

@app.route('/find_term', methods=['POST'])
def find_term():
    data = request.get_json()
    word = data['word']
    phrase = data['phrase']

    # Process the phrase with spaCy
    doc = nlp(phrase)

    # Find the token that stems from the given word
    for token in doc:
        if token.lemma_ == word:
            return jsonify({
                'start_index': token.idx,
                'end_index': token.idx + len(token.text)
            })

    # If no match is found
    return jsonify({'error': 'Term not found in the phrase'}), 404

if __name__ == '__main__':
    app.run(debug=True)
