import { readFileSync } from 'fs';
import { resolve } from 'path';
import * as semver from 'semver';
import * as ts from 'typescript';

export interface TransformerOptions {}

const entry = resolve(process.cwd(), 'scripts/e2e-selectors/components.ts');
const sourceFile = ts.createSourceFile(
  'components.ts',
  readFileSync(entry).toString(),
  ts.ScriptTarget.ES2015,
  /*setParentNodes */ true
);

const version = '8.6.0';

const getInitializedForVersion = (properties: ts.NodeArray<ts.ObjectLiteralElementLike>): ts.PropertyAssignment => {
  let current: ts.PropertyAssignment;
  for (const property of properties) {
    if (ts.isStringLiteral(property.name)) {
      if (ts.isPropertyAssignment(property) && semver.satisfies(version, `>=${property.name.text}`)) {
        if (!current) {
          current = property;
        } else if (semver.gt(property.name.text, current.name.getText())) {
          current = property;
        }
      }
    }
  }

  return current;
};

const replaceVersions = (context: ts.TransformationContext) => (rootNode: ts.Node) => {
  const visit = (node: ts.Node): ts.Node => {
    const newNode = ts.visitEachChild(node, visit, context);
    if (ts.isObjectLiteralExpression(newNode)) {
      const propertyAssignment = getInitializedForVersion(newNode.properties);
      if (!propertyAssignment) {
        return newNode;
      }

      if (ts.isStringLiteral(propertyAssignment.name) && ts.isStringLiteral(propertyAssignment.initializer)) {
        return propertyAssignment.initializer;
      } else if (ts.isStringLiteral(propertyAssignment.name) && ts.isArrowFunction(propertyAssignment.initializer)) {
        return propertyAssignment.initializer;
      }
    }

    return newNode;
  };

  return ts.visitNode(rootNode, visit);
};

const transformationResult = ts.transform(sourceFile, [replaceVersions]);
const printer = ts.createPrinter({ newLine: ts.NewLineKind.LineFeed });

console.log(
  printer.printNode(
    ts.EmitHint.Unspecified,
    transformationResult.transformed[0],
    ts.createSourceFile('', '', ts.ScriptTarget.Latest)
  )
);

const init = async () => {
  const entry = resolve(process.cwd(), 'scripts/e2e-selectors/components.ts');
  const program = ts.createProgram([entry], {
    allowSyntheticDefaultImports: true,
  });
  const sourceFile = program.getSourceFile(entry);
  const selectorRootNode = getSelectorsRootNode(sourceFile!);
  if (!selectorRootNode) {
    console.error('Could not find selectors root node');
    process.exit(1);
  }

  if (!ts.isObjectLiteralExpression(selectorRootNode.initializer)) {
    console.error('Could not find selectors root node with object litteral');
    process.exit(1);
  }

  pumpUpTheMusic(selectorRootNode.initializer);
};

function pumpUpTheMusic(exp: ts.ObjectLiteralExpression) {
  // I think we need to break the loop when we have a match
  // since we are replacing parts of the tree...
  exp.forEachChild((node) => {
    if (ts.isPropertyAssignment(node)) {
      if (ts.isObjectLiteralExpression(node.initializer)) {
        return pumpUpTheMusic(node.initializer);
      }
      if (ts.isArrowFunction(node.initializer) && ts.isStringLiteral(node.name)) {
        return replaceChildrenWithSelector(exp);
      }
      if (ts.isStringLiteral(node.initializer) && ts.isStringLiteral(node.name)) {
        return replaceChildrenWithSelector(exp);
      }
    }
  });
}

function replaceChildrenWithSelector(ps: ts.ObjectLiteralExpression) {
  const [child] = ps.properties;
  if (ts.isPropertyAssignment(child)) {
    if (ts.isStringLiteral(child.name)) {
      console.log('replcae', child.name.text);
    }
  }
}

const getSelectorsRootNode = (sourceFile: ts.SourceFile): ts.VariableDeclaration | null => {
  let rootNode: ts.VariableDeclaration | null = null;

  ts.forEachChild(sourceFile, (node) => {
    if (!ts.isVariableStatement(node)) {
      return;
    }

    node.declarationList.declarations.forEach((declaration) => {
      if (!ts.isVariableDeclaration(declaration)) {
        return;
      }
      if (!ts.isIdentifier(declaration.name)) {
        return;
      }
      if (declaration.name.escapedText !== 'components') {
        return;
      }
      rootNode = declaration;
    });
  });

  return rootNode;
};

// init();