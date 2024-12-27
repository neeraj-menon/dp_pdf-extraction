import sys
from pdf2image import convert_from_path
import os

def convert_pdf_to_images(pdf_path, output_dir):
    try:
        # Convert PDF pages to images
        pages = convert_from_path(pdf_path, dpi=300)
        
        # Save each page as a JPEG file
        for i, page in enumerate(pages, start=1):
            output_path = os.path.join(output_dir, f'page_{i}.jpg')
            page.save(output_path, 'JPEG')
            print(output_path)  # Print path for Go program to capture
            
    except Exception as e:
        print(f"Error: {str(e)}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python pdf_to_image.py <pdf_path> <output_dir>", file=sys.stderr)
        sys.exit(1)
        
    pdf_path = sys.argv[1]
    output_dir = sys.argv[2]
    
    convert_pdf_to_images(pdf_path, output_dir)
