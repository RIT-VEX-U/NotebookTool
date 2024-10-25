const PRINT_OPTIONS = {
  clearCache: true,
  printOptions: {
    scale: 0.6
  },

  completionTrigger: new HtmlPdf.CompletionTrigger.Timer(5000)  // Give it 5000ms to render the HTML
};

async function outputHTMLToPDF(sourceHTML, outputFilename) {
  console.log("Printing the html using Chrome...");
  let pdf = await HtmlPdf.create(sourceHTML, PRINT_OPTIONS);

  console.log("Saving the PDF to " + outputFilename + "...");
  await pdf.toFile(path.join(DEFAULT_PRINT_PATH, outputFilename));
}
